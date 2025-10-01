package music

import (
	"fmt"
	"strconv"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

func (ms *MusicService) HandlePlay(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, _, vc := gl.GetVoiceChannelID(ms.session, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	if len(args) == 0 {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	voice, err := ms.GetVoiceConnection(vc, ms.session, m.GuildID)
	if err != nil {
		return gl.EmbedMessage(err.Error())
	}

	q := ms.GetOrCreateQueue(voice, vc)

	opt := miri.SearchOptions{
		Limit: 1,
		Query: strings.Join(args, " "),
	}
	results, err := ms.Client.SearchTracks(ms.Ctx, opt)
	if err != nil {
		ms.Logger.Errorf("could not search track: %v", err)
		if q.nowPlaying == nil {
			voice.Disconnect(ms.Ctx)
		}
		return gl.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		if q.nowPlaying == nil {
			voice.Disconnect(ms.Ctx)
		}
		return gl.EmbedMessage(gl.MsgNoResults)
	}

	track := &results[0]
	q.AddTrack(ms, track)

	return gl.EmbedTrackMessage(track)
}

func (ms *MusicService) HandleSearch(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := strings.Join(args, " ")
	if q == "" {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	opt := miri.SearchOptions{
		Limit: ms.maxSearchResults,
		Query: q,
	}
	results, err := ms.Client.SearchTracks(ms.Ctx, opt)
	if err != nil {
		ms.Logger.Errorf("could not search track: %v", err)
		return gl.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		return gl.EmbedMessage(gl.MsgNoResults)
	}

	maxResults := min(len(results), int(ms.maxSearchResults))
	var out string
	var buttons []discordgo.MessageComponent

	for i := range maxResults {
		v := results[i]
		out += fmt.Sprintf(gl.MsgOrderedList, i+1, gl.FormatTrackLine(&v))

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%d", i+1),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("choose_track:%d", i+1),
		})
	}

	// add cancel button
	buttons = append(buttons, discordgo.Button{
		Label:    "Cancel",
		Style:    discordgo.DangerButton,
		CustomID: "choose_track:0",
	})

	key := gl.GetPendingSearchKey(m.ChannelID, m.Author.ID)
	ms.Searches[key] = results[:maxResults]

	// Split buttons into rows of max 5
	var components []discordgo.MessageComponent
	for i := 0; i < len(buttons); i += 5 {
		end := min(i+5, len(buttons))
		row := discordgo.ActionsRow{
			Components: buttons[i:end],
		}
		components = append(components, row)
	}

	msg := gl.EmbedMessage(out)
	msg.Components = components
	return msg
}

func (ms *MusicService) HandleLyrics(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := ms.GetQueue(m.GuildID)
	if q == nil || q.nowPlaying == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	lyrics, err := q.nowPlaying.Lyrics()
	if err != nil || lyrics == "" {
		ms.Logger.Errorf("could not fetch lyrics: %v", err)
		return gl.EmbedMessage(gl.MsgNoLyrics)
	}

	if len(lyrics) > gl.DiscordEmbedDescriptionLimit { // quick bytes check
		runes := []rune(lyrics)
		if len(runes) > gl.DiscordEmbedDescriptionLimit { // accurate rune check
			lyrics = string(runes[:gl.DiscordEmbedDescriptionLimit-1]) + "…"
		}
	}

	response := gl.EmbedTrackMessage(q.nowPlaying)
	response.Embeds[0].Description = lyrics

	return response
}

func (ms *MusicService) HandleSkip(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(ms.session, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return gl.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	err := q.PlayNext(ms, true)
	if err != nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	return gl.EmbedMessage(gl.MsgSkipped)
}

func (ms *MusicService) HandleQueue(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := ms.GetQueue(m.GuildID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	var out string
	tracks := q.Tracks()
	for i, v := range tracks {
		out += fmt.Sprintf(gl.MsgOrderedList, i, gl.FormatTrackLine(&v))
	}
	return gl.EmbedMessage(out)
}

func (ms *MusicService) HandleClear(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(ms.session, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return gl.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	q.Clear()

	return gl.EmbedMessage(gl.MsgCleared)
}

func (ms *MusicService) HandleLeave(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(ms.session, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return gl.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	ms.DeleteQueue(g.ID)
	return gl.EmbedMessage(gl.MsgLeft)
}

func (ms *MusicService) HandleChooseTrack(arg string, i *discordgo.InteractionCreate) *discordgo.MessageSend {
	trackIdx, err := strconv.Atoi(arg)
	if err != nil || trackIdx < 0 {
		return gl.EmbedMessage(gl.MsgInvalidTrackNumber)
	}

	key := gl.GetPendingSearchKey(i.ChannelID, i.Member.User.ID)
	results, found := ms.Searches[key]
	if !found || trackIdx > len(results) {
		return gl.EmbedMessage(gl.MsgCantFindSearch)
	}

	if trackIdx == 0 {
		// Cancel selection silently
		delete(ms.Searches, key)
		ms.session.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		return nil
	}

	track := &results[trackIdx-1]
	r, _, vc := gl.GetVoiceChannelID(ms.session, i.Member, i.GuildID, i.Member.User.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	voice, err := ms.GetVoiceConnection(vc, ms.session, i.GuildID)
	if err != nil {
		return gl.EmbedMessage(err.Error())
	}

	q := ms.GetOrCreateQueue(voice, vc)
	q.AddTrack(ms, track)
	delete(ms.Searches, key)
	defer ms.session.ChannelMessageDelete(i.ChannelID, i.Message.ID)

	return gl.EmbedTrackMessage(track)
}

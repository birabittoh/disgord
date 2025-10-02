package music

import (
	"fmt"
	"strconv"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

func getPendingSearchKey(channelID, authorID string) string {
	return channelID + ":" + authorID
}

func (ms *MusicService) PlayToVC(query string, vc string, guildID string) (response string, track *miri.SongResult, err error) {
	voice, err := ms.GetVoiceConnection(vc, guildID)
	if err != nil {
		return
	}

	q := ms.GetOrCreateQueue(voice, vc)

	opt := miri.SearchOptions{
		Limit: 1,
		Query: query,
	}
	results, err := ms.Client.SearchTracks(ms.us.Ctx, opt)
	if err != nil {
		ms.Logger.Errorf("could not search track: %v", err)
		if q.nowPlaying == nil {
			voice.Disconnect(ms.us.Ctx)
		}
		response = gl.MsgError
		return
	}

	if len(results) == 0 {
		if q.nowPlaying == nil {
			voice.Disconnect(ms.us.Ctx)
		}
		response = gl.MsgNoResults
		return
	}

	track = &results[0]
	q.AddTrack(ms, track)
	return
}

func (ms *MusicService) HandlePlay(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, _, vc := ms.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	if len(args) == 0 {
		return ms.us.EmbedMessage(gl.MsgNoKeywords)
	}

	query := strings.Join(args, " ")
	response, track, err := ms.PlayToVC(query, vc, m.GuildID)
	if err != nil {
		return ms.us.EmbedMessage(gl.MsgError)
	}

	if track != nil {
		return ms.us.EmbedTrackMessage(track)
	}

	return ms.us.EmbedMessage(response)
}

func (ms *MusicService) HandleSearch(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := strings.Join(args, " ")
	if q == "" {
		return ms.us.EmbedMessage(gl.MsgNoKeywords)
	}

	opt := miri.SearchOptions{
		Limit: uint64(ms.us.Config.MaxSearchResults),
		Query: q,
	}
	results, err := ms.Client.SearchTracks(ms.us.Ctx, opt)
	if err != nil {
		ms.Logger.Errorf("could not search track: %v", err)
		return ms.us.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		return ms.us.EmbedMessage(gl.MsgNoResults)
	}

	maxResults := min(len(results), int(ms.us.Config.MaxSearchResults))
	var out string
	var buttons []discordgo.MessageComponent

	for i := range maxResults {
		v := results[i]
		out += fmt.Sprintf(gl.MsgOrderedList, i+1, ms.us.FormatTrackLine(&v))

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

	key := getPendingSearchKey(m.ChannelID, m.Author.ID)
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

	msg := ms.us.EmbedMessage(out)
	msg.Components = components
	return msg
}

func (ms *MusicService) HandleLyrics(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := ms.GetQueue(m.GuildID)
	if q == nil || q.nowPlaying == nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	lyrics, err := q.nowPlaying.Lyrics()
	if err != nil || lyrics == "" {
		ms.Logger.Errorf("could not fetch lyrics: %v", err)
		return ms.us.EmbedMessage(gl.MsgNoLyrics)
	}

	if len(lyrics) > gl.DiscordEmbedDescriptionLimit { // quick bytes check
		runes := []rune(lyrics)
		if len(runes) > gl.DiscordEmbedDescriptionLimit { // accurate rune check
			lyrics = string(runes[:gl.DiscordEmbedDescriptionLimit-1]) + "…"
		}
	}

	response := ms.us.EmbedTrackMessage(q.nowPlaying)
	response.Embeds[0].Description = lyrics

	return response
}

func (ms *MusicService) HandleSkip(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := ms.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return ms.us.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	err := q.PlayNext(ms, true)
	if err != nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	return ms.us.EmbedMessage(gl.MsgSkipped)
}

func (ms *MusicService) HandleQueue(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := ms.GetQueue(m.GuildID)
	if q == nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	var out string
	tracks := q.Tracks()
	for i, v := range tracks {
		out += fmt.Sprintf(gl.MsgOrderedList, i, ms.us.FormatTrackLine(&v))
	}
	return ms.us.EmbedMessage(out)
}

func (ms *MusicService) HandleClear(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := ms.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return ms.us.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	q.Clear()

	return ms.us.EmbedMessage(gl.MsgCleared)
}

func (ms *MusicService) HandleLeave(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := ms.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	q := ms.GetQueue(g.ID)
	if q == nil {
		return ms.us.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return ms.us.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	ms.DeleteQueue(g.ID)
	return ms.us.EmbedMessage(gl.MsgLeft)
}

func (ms *MusicService) HandleChooseTrack(arg string, i *discordgo.InteractionCreate) *discordgo.MessageSend {
	trackIdx, err := strconv.Atoi(arg)
	if err != nil || trackIdx < 0 {
		return ms.us.EmbedMessage(gl.MsgInvalidTrackNumber)
	}

	key := getPendingSearchKey(i.ChannelID, i.Member.User.ID)
	results, found := ms.Searches[key]
	if !found || trackIdx > len(results) {
		return ms.us.EmbedMessage(gl.MsgCantFindSearch)
	}

	if trackIdx == 0 {
		// Cancel selection silently
		delete(ms.Searches, key)
		ms.us.Session.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		return nil
	}

	track := &results[trackIdx-1]
	r, _, vc := ms.us.GetVoiceChannelID(i.Member, i.GuildID, i.Member.User.ID)
	if r != "" {
		return ms.us.EmbedMessage(r)
	}

	voice, err := ms.GetVoiceConnection(vc, i.GuildID)
	if err != nil {
		return ms.us.EmbedMessage(err.Error())
	}

	q := ms.GetOrCreateQueue(voice, vc)
	q.AddTrack(ms, track)
	delete(ms.Searches, key)
	defer ms.us.Session.ChannelMessageDelete(i.ChannelID, i.Message.ID)

	return ms.us.EmbedTrackMessage(track)
}

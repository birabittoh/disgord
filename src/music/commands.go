package music

import (
	"fmt"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

func (ms *MusicService) HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, _, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	if len(args) == 0 {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	voice, err := ms.GetVoiceConnection(vc, s, m.GuildID)
	if err != nil {
		return gl.EmbedMessage(err.Error())
	}

	q := ms.GetOrCreateQueue(voice, vc)

	query := strings.Join(args, " ")
	results, err := ms.Client.SearchTracks(ms.Ctx, query)
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

	coverURL := track.CoverURL(gl.AlbumCoverSize)
	return gl.EmbedTrackMessage(gl.FormatTrack(track), coverURL)
}

func (ms *MusicService) HandleSearch(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := strings.Join(args, " ")
	if q == "" {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	results, err := ms.Client.SearchTracks(ms.Ctx, q)
	if err != nil {
		ms.Logger.Errorf("could not search track: %v", err)
		return gl.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		return gl.EmbedMessage(gl.MsgNoResults)
	}

	maxResults := min(len(results), 9)
	var out string
	var buttons []discordgo.MessageComponent

	for i := range maxResults {
		v := results[i]
		out += fmt.Sprintf(gl.MsgOrderedList, i+1, gl.FormatTrackLine(&v))

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%d", i+1),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("choose_track_%d", i+1),
		})
	}

	// add cancel button
	buttons = append(buttons, discordgo.Button{
		Label:    "Cancel",
		Style:    discordgo.DangerButton,
		CustomID: "choose_track_0",
	})

	out += gl.MsgSearchHelp

	key := gl.GetPendingSearchKey(m.ChannelID, m.Author.ID)
	gl.PendingSearches[key] = results[:maxResults]

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

func (ms *MusicService) HandleSkip(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
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

func (ms *MusicService) HandleQueue(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
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

func (ms *MusicService) HandleClear(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
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

func (ms *MusicService) HandleLeave(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
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

	err := q.Stop()
	if err != nil {
		return gl.EmbedMessage(gl.MsgError)
	}

	return gl.EmbedMessage(gl.MsgLeft)
}

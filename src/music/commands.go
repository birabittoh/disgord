package music

import (
	"context"
	"fmt"
	"strings"
	"time"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/birabittoh/miri/deezer"
	"github.com/bwmarrin/discordgo"
)

const maxResultsAmount = 9

var d *miri.Client

var mainCtx *context.Context

func Init(ctx *context.Context) error {
	mainCtx = ctx
	cfg, err := deezer.NewConfig(gl.Config.Values.ArlCookie, gl.Config.Values.SecretKey)
	if err != nil {
		return err
	}
	cfg.Timeout = 30 * time.Minute // long timeout for music streaming
	d, err = miri.New(*ctx, cfg)
	return err
}

func GetVoiceConnection(vc string, s *discordgo.Session, guildID string) (voice *discordgo.VoiceConnection, err error) {
	alreadyConnected := false
	for _, vs := range s.VoiceConnections {
		if vs.GuildID == guildID {
			voice = vs
			alreadyConnected = true
			break
		}
	}
	if !alreadyConnected {
		var err error
		voice, err = s.ChannelVoiceJoin(*mainCtx, guildID, vc, false, true)
		if err != nil {
			logger.Errorf("could not join voice channel: %v", err)
			return nil, err
		}
	}

	return voice, nil
}

func HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, _, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	if len(args) == 0 {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	voice, err := GetVoiceConnection(vc, s, m.GuildID)
	if err != nil {
		return gl.EmbedMessage(err.Error())
	}

	// Get the queue for the guild
	q := GetOrCreateQueue(voice, vc)

	query := strings.Join(args, " ")
	results, err := d.SearchTracks(*mainCtx, query)
	if err != nil {
		logger.Errorf("could not search track: %v", err)
		if q.nowPlaying == nil {
			voice.Disconnect(*mainCtx)
		}
		return gl.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		if q.nowPlaying == nil {
			voice.Disconnect(*mainCtx)
		}
		return gl.EmbedMessage(gl.MsgNoResults)
	}

	track := &results[0]

	// Add track to the queue
	q.AddTrack(track)

	coverURL := track.CoverURL(gl.AlbumCoverSize)
	return gl.EmbedTrackMessage(gl.FormatTrack(track), coverURL)
}

func HandleSearch(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := strings.Join(args, " ")
	if q == "" {
		return gl.EmbedMessage(gl.MsgNoKeywords)
	}

	results, err := d.SearchTracks(*mainCtx, q)
	if err != nil {
		logger.Errorf("could not search track: %v", err)
		return gl.EmbedMessage(gl.MsgError)
	}

	if len(results) == 0 {
		return gl.EmbedMessage(gl.MsgNoResults)
	}

	maxResults := min(len(results), maxResultsAmount)
	var out string
	var buttons []discordgo.MessageComponent

	for i := range maxResults {
		v := results[i]
		duration := time.Duration(v.Duration) * time.Second
		out += fmt.Sprintf(gl.MsgSearchLine, i+1, gl.FormatTrackLine(&v), duration.String())

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%d", i+1),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("choose_track_%d", i+1),
		})
	}

	out += gl.MsgSearchHelp

	key := gl.GetPendingSearchKey(m.ChannelID, m.Author.ID)
	gl.PendingSearches[key] = results[:maxResults]

	// Split buttons into rows of max 5
	var components []discordgo.MessageComponent
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		row := discordgo.ActionsRow{
			Components: buttons[i:end],
		}
		components = append(components, row)
	}

	msg := gl.EmbedMessage(out)
	msg.Components = components
	return msg
}

/*

func HandlePause(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return MsgSameVoiceChannel
	}

	q.Pause()

	return MsgPaused
}

func HandleResume(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return MsgSameVoiceChannel
	}

	q.Resume()

	return MsgResumed
}

*/

func HandleSkip(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := GetQueue(g.ID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return gl.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	err := q.PlayNext(true)
	if err != nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	return gl.EmbedMessage(gl.MsgNothingIsPlaying)
}

func HandleQueue(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	q := GetQueue(m.GuildID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	var out string
	tracks := q.Tracks()
	for i, v := range tracks {
		out += fmt.Sprintf(gl.MsgQueueLine, i, gl.FormatTrackLine(&v))
	}
	return gl.EmbedMessage(out)
}

func HandleClear(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := GetQueue(g.ID)
	if q == nil {
		return gl.EmbedMessage(gl.MsgNothingIsPlaying)
	}

	if vc != q.VoiceChannelID() {
		return gl.EmbedMessage(gl.MsgSameVoiceChannel)
	}

	q.Clear()

	return gl.EmbedMessage(gl.MsgCleared)
}

func HandleLeave(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	r, g, vc := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if r != "" {
		return gl.EmbedMessage(r)
	}

	q := GetQueue(g.ID)
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

func HandleBotVSU(vsu *discordgo.VoiceStateUpdate) {
	if vsu.BeforeUpdate == nil {
		// user joined a voice channel
		return
	}

	queue := GetQueue(vsu.GuildID)
	if queue == nil {
		// no queue for this guild
		return
	}

	if queue.NowPlaying() == nil {
		// song has ended naturally
		return
	}

	vc := queue.VoiceConnection()
	if vc == nil {
		return
	}

	if vsu.ChannelID == "" && vsu.BeforeUpdate.ChannelID == queue.VoiceChannelID() {
		logger.Println("Bot disconnected from voice channel, stopping audio playback.")
		queue.Stop()
	}
}

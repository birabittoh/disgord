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

func HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, _, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	if len(args) == 0 {
		return gl.MsgNoKeywords
	}

	var voice *discordgo.VoiceConnection
	alreadyConnected := false
	for _, vs := range s.VoiceConnections {
		if vs.GuildID == m.GuildID {
			voice = vs
			alreadyConnected = true
			break
		}
	}
	if !alreadyConnected {
		var err error
		voice, err = s.ChannelVoiceJoin(*mainCtx, m.GuildID, vc, false, true)
		if err != nil {
			logger.Errorf("could not join voice channel: %v", err)
			return gl.MsgError
		}
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
		return gl.MsgError
	}

	if len(results) == 0 {
		if q.nowPlaying == nil {
			voice.Disconnect(*mainCtx)
		}
		return gl.MsgNoResults
	}

	track := &results[0]

	// Add track to the queue
	q.AddTrack(track)

	return fmt.Sprintf(gl.MsgAddedToQueue, gl.FormatTrack(track))
}

func HandleSearch(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	q := strings.Join(args, " ")
	if q == "" {
		return gl.MsgNoKeywords
	}

	results, err := d.SearchTracks(*mainCtx, q)
	if err != nil {
		logger.Errorf("could not search track: %v", err)
		return gl.MsgError
	}

	if len(results) == 0 {
		return gl.MsgNoResults
	}

	var out string
	maxResults := min(len(results), 5)
	for i := 0; i < maxResults; i++ {
		v := results[i]
		duration := time.Duration(v.Duration) * time.Second
		out += fmt.Sprintf(gl.MsgSearchLine, i+1, gl.FormatTrack(&v), duration.String())
	}
	return out
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

func HandleSkip(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return gl.MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return gl.MsgSameVoiceChannel
	}

	err := q.PlayNext(true)
	if err != nil {
		return gl.MsgNothingIsPlaying
	}

	return gl.MsgSkipped
}

func HandleQueue(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	q := GetQueue(m.GuildID)
	if q == nil {
		return gl.MsgNothingIsPlaying
	}

	var out string
	tracks := q.Tracks()
	for i, v := range tracks {
		out += fmt.Sprintf(gl.MsgQueueLine, i, gl.FormatTrack(&v))
	}
	return out
}

func HandleClear(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return gl.MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return gl.MsgSameVoiceChannel
	}

	q.Clear()

	return gl.MsgCleared
}

func HandleLeave(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return gl.MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return gl.MsgSameVoiceChannel
	}

	err := q.Stop()
	if err != nil {
		return gl.MsgError
	}

	return gl.MsgLeft
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

package music

import (
	"context"
	"fmt"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/birabittoh/miri/deezer"
	"github.com/bwmarrin/discordgo"
)

const (
	MsgNoURL            = "Please, provide a YouTube URL."
	MsgAddedToQueue     = "Added to queue: %s."
	MsgNothingIsPlaying = "Nothing is playing."
	MsgSameVoiceChannel = "You need to be in the same voice channel to use this command."
	MsgPaused           = "Paused."
	MsgResumed          = "Resumed."
	MsgSkipped          = "Skipped."
	MsgCleared          = "Cleared."
	MsgLeft             = "Left."
	MsgQueueLine        = "%d. %s\n"
)

var d *miri.Client

var mainCtx *context.Context

func Init(ctx *context.Context) error {
	mainCtx = ctx
	cfg, err := deezer.NewConfig(gl.Config.Values.ArlCookie, gl.Config.Values.SecretKey)
	if err != nil {
		return err
	}
	d, err = miri.New(*ctx, cfg)
	return err
}

func HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, _, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	if len(args) == 0 {
		return MsgNoURL
	}

	voice, err := s.ChannelVoiceJoin(*mainCtx, m.GuildID, vc, false, true)
	if err != nil {
		logger.Errorf("could not join voice channel: %v", err)
		return gl.MsgError
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

	return fmt.Sprintf(MsgAddedToQueue, gl.FormatTrack(track))
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
		return MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return MsgSameVoiceChannel
	}

	err := q.PlayNext(true)
	if err != nil {
		return MsgNothingIsPlaying
	}

	return MsgSkipped
}

func HandleQueue(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	q := GetQueue(m.GuildID)
	if q == nil {
		return MsgNothingIsPlaying
	}

	var out string
	tracks := q.Tracks()
	for i, v := range tracks {
		out += fmt.Sprintf(MsgQueueLine, i, gl.FormatTrack(&v))
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
		return MsgNothingIsPlaying
	}

	if vc != q.VoiceChannelID() {
		return MsgSameVoiceChannel
	}

	q.Clear()

	return MsgCleared
}

func HandleLeave(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
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

	err := q.Stop()
	if err != nil {
		return gl.MsgError
	}

	return MsgLeft
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

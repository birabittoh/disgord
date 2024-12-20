package music

import (
	"fmt"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/rabbitpipe"
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

var yt *rabbitpipe.Client

func Init(instance string) {
	yt = rabbitpipe.New(instance)
}

func HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, _, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	if len(args) == 0 {
		return MsgNoURL
	}

	voice, err := s.ChannelVoiceJoin(m.GuildID, vc, false, true)
	if err != nil {
		logger.Errorf("could not join voice channel: %v", err)
		return gl.MsgError
	}

	// Get the queue for the guild
	q := GetOrCreateQueue(voice)

	// Get the video information
	video, err := getVideo(args)
	if err != nil {
		logger.Errorf("could not get video: %v", err)
		if q.nowPlaying == nil {
			voice.Disconnect()
		}
		return gl.MsgError
	}

	// Add video to the queue
	q.AddVideo(video)

	return fmt.Sprintf(MsgAddedToQueue, gl.FormatVideo(video))
}

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

	err := q.PlayNext()
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
	videos := q.Videos()
	for i, v := range videos {
		out += fmt.Sprintf(MsgQueueLine, i, gl.FormatVideo(v))
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

	if vsu.ChannelID == "" && vsu.BeforeUpdate.ChannelID == vc.ChannelID {
		logger.Println("Bot disconnected from voice channel, stopping audio playback.")
		queue.Stop()
	}
}

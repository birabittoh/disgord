package music

import (
	gl "github.com/BiRabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

func HandlePlay(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, _, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	if len(args) == 0 {
		return "Please, provide a YouTube URL."
	}

	voice, err := s.ChannelVoiceJoin(m.GuildID, vc, false, true)
	if err != nil {
		logger.Errorf("could not join voice channel: %v", err)
		return gl.MsgError
	}

	// Get the video information
	video, err := gl.YT.GetVideo(args[0])
	if err != nil {
		logger.Errorf("could not get video: %v", err)
		return gl.MsgError
	}

	// Get the queue for the guild
	q := GetOrCreateQueue(voice)

	// Add video to the queue
	q.AddVideo(video)

	return "Added to queue: " + gl.FormatVideo(video)
}

func HandlePause(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return "Nothing is playing."
	}

	if vc != q.VoiceChannelID() {
		return "You need to be in the same voice channel to use this command."
	}

	q.Pause()

	return "Paused."
}

func HandleResume(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return "Nothing is playing."
	}

	if vc != q.VoiceChannelID() {
		return "You need to be in the same voice channel to use this command."
	}

	q.Resume()

	return "Resumed."
}

func HandleSkip(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return "Nothing is playing."
	}

	if vc != q.VoiceChannelID() {
		return "You need to be in the same voice channel to use this command."
	}

	err := q.PlayNext()
	if err != nil {
		return "Nothing is playing."
	}

	return "Skipped."
}

func HandleQueue(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	q := GetQueue(m.GuildID)
	if q == nil {
		return "Nothing is playing."
	}

	var out string
	videos := q.Videos()
	for _, v := range videos {
		out += gl.FormatVideo(v) + "\n"
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
		return "Nothing is playing."
	}

	if vc != q.VoiceChannelID() {
		return "You need to be in the same voice channel to use this command."
	}

	q.Clear()

	return "Cleared."
}

func HandleLeave(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	r, g, vc := gl.GetVoiceChannelID(s, m)
	if r != "" {
		return r
	}

	q := GetQueue(g.ID)
	if q == nil {
		return "Nothing is playing."
	}

	if vc != q.VoiceChannelID() {
		return "You need to be in the same voice channel to use this command."
	}

	err := q.Stop()
	if err != nil {
		return gl.MsgError
	}

	return "Cleared."
}

package music

import (
	"os"

	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

var logger = mylog.NewLogger(os.Stdin, "music", mylog.DEBUG)

type Queue struct {
	nowPlaying  string
	items       []string
	audioStream *Audio
	vc          *discordgo.VoiceConnection
}

// queues stores all guild queues
var queues = map[string]*Queue{}

// GetOrCreateQueue fetches or creates a new queue for the guild
func GetOrCreateQueue(vc *discordgo.VoiceConnection) (q *Queue) {
	q, ok := queues[vc.GuildID]
	if !ok {
		q = &Queue{vc: vc}
		queues[vc.GuildID] = q
		return
	}

	if !q.vc.Ready {
		q.vc = vc
	}
	return
}

// GetQueue returns either nil or the queue for the requested guild
func GetQueue(guildID string) *Queue {
	q, ok := queues[guildID]
	if ok {
		return q
	}
	return nil
}

// AddVideo adds a new video to the queue
func (q *Queue) AddVideo(url string) {
	q.AddVideos([]string{url})
}

// AddVideos adds a list of videos to the queue
func (q *Queue) AddVideos(videos []string) {
	q.items = append(q.items, videos...)
	if q.nowPlaying == "" {
		err := q.PlayNext(false)
		if err != nil {
			logger.Error(err)
		}
	}
}

// PlayNext starts playing the next video in the queue
func (q *Queue) PlayNext(skip bool) (err error) {
	if q.audioStream != nil && q.audioStream.playing {
		q.audioStream.Stop()

		// If this is a manual skip, return early to avoid recursion
		// The monitor callback will handle playing the next song
		if skip {
			return nil
		}
	}

	if len(q.items) == 0 {
		q.nowPlaying = ""
		return q.vc.Disconnect()
	}

	q.nowPlaying = q.items[0]
	q.items = q.items[1:]

	q.audioStream, err = NewAudio(q.nowPlaying, q.vc)
	if err != nil {
		return
	}

	q.audioStream.Monitor(func() { q.PlayNext(false) })
	return
}

// Stop stops the player and clears the queue
func (q *Queue) Stop() error {
	q.Clear()
	q.nowPlaying = ""

	if q.audioStream != nil {
		q.audioStream.Stop()
	}

	if q.vc != nil {
		return q.vc.Disconnect()
	}

	return nil
}

// Clear clears the video queue
func (q *Queue) Clear() {
	q.items = []string{}
}

// Videos returns all videos in the queue including the now playing one
func (q *Queue) Videos() []string {
	if q.nowPlaying != "" {
		return append([]string{q.nowPlaying}, q.items...)
	}
	return q.items
}

func (q *Queue) VoiceChannelID() string {
	return q.vc.ChannelID
}

func (q *Queue) AudioStream() *Audio {
	return q.audioStream
}

func (q *Queue) VoiceConnection() *discordgo.VoiceConnection {
	return q.vc
}

func (q *Queue) NowPlaying() string {
	return q.nowPlaying
}

package music

import (
	"os"

	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

var logger = mylog.NewLogger(os.Stdin, "music", mylog.DEBUG)

type Queue struct {
	nowPlaying  *youtube.Video
	items       []*youtube.Video
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
func (q *Queue) AddVideo(video *youtube.Video) {
	q.items = append(q.items, video)
	if q.nowPlaying == nil {
		q.PlayNext()
	}
}

// AddVideos adds a list of videos to the queue
func (q *Queue) AddVideos(videos []*youtube.Video) {
	q.items = append(q.items, videos...)
	if q.nowPlaying == nil {
		err := q.PlayNext()
		if err != nil {
			logger.Error(err)
		}
	}
}

// PlayNext starts playing the next video in the queue
func (q *Queue) PlayNext() (err error) {
	if q.audioStream != nil {
		q.audioStream.Stop()
	}

	if len(q.items) == 0 {
		q.nowPlaying = nil
		return q.vc.Disconnect()
	}

	q.nowPlaying = q.items[0]
	q.items = q.items[1:]

	format := getFormat(*q.nowPlaying)
	if format == nil {
		logger.Debug("no formats with audio channels available for video " + q.nowPlaying.ID)
		return q.PlayNext()
	}

	q.audioStream, err = NewAudio(format.URL, q.vc)
	if err != nil {
		return
	}

	q.audioStream.Monitor(func() { q.PlayNext() })
	return
}

// Stop stops the player and clears the queue
func (q *Queue) Stop() error {
	q.Clear()
	q.nowPlaying = nil

	if q.audioStream != nil {
		q.audioStream.Stop()
	}

	if q.vc != nil {
		return q.vc.Disconnect()
	}

	return nil
}

// Pause pauses the player
func (q *Queue) Pause() {
	q.audioStream.Pause()
}

// Resume resumes the player
func (q *Queue) Resume() {
	q.audioStream.Resume()
}

// Clear clears the video queue
func (q *Queue) Clear() {
	q.items = []*youtube.Video{}
}

// Videos returns all videos in the queue including the now playing one
func (q *Queue) Videos() []*youtube.Video {
	if q.nowPlaying != nil {
		return append([]*youtube.Video{q.nowPlaying}, q.items...)
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

func (q *Queue) NowPlaying() *youtube.Video {
	return q.nowPlaying
}

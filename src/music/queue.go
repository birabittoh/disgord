package music

import (
	"os"

	"github.com/birabittoh/disgord/src/mylog"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

var logger = mylog.NewLogger(os.Stdin, "music", mylog.DEBUG)

type Queue struct {
	nowPlaying  *miri.SongResult
	items       []miri.SongResult
	audioStream *Audio
	vc          *discordgo.VoiceConnection
	channelID   string
}

// queues stores all guild queues
var queues = map[string]*Queue{}

// GetOrCreateQueue fetches or creates a new queue for the guild
func GetOrCreateQueue(vc *discordgo.VoiceConnection, channelID string) (q *Queue) {
	q, ok := queues[vc.GuildID]
	if !ok {
		q = &Queue{}
		queues[vc.GuildID] = q
	}

	q.vc = vc
	q.channelID = channelID
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

// AddTrack adds a new track to the queue
func (q *Queue) AddTrack(track *miri.SongResult) {
	q.AddTracks([]miri.SongResult{*track})
}

// AddTracks adds a list of tracks to the queue
func (q *Queue) AddTracks(tracks []miri.SongResult) {
	q.items = append(q.items, tracks...)
	if q.nowPlaying == nil {
		err := q.PlayNext(false)
		if err != nil {
			logger.Error(err)
		}
	}
}

// PlayNext starts playing the next track in the queue
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
		q.nowPlaying = nil
		return q.vc.Disconnect(*mainCtx)
	}

	q.nowPlaying = &q.items[0]
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
	q.nowPlaying = nil

	if q.audioStream != nil {
		q.audioStream.Stop()
	}

	if q.vc != nil {
		return q.vc.Disconnect(*mainCtx)
	}

	return nil
}

// Clear clears the track queue
func (q *Queue) Clear() {
	q.items = []miri.SongResult{}
}

// Tracks returns all tracks in the queue including the now playing one
func (q *Queue) Tracks() []miri.SongResult {
	if q.nowPlaying != nil {
		return append([]miri.SongResult{*q.nowPlaying}, q.items...)
	}
	return q.items
}

func (q *Queue) VoiceChannelID() string {
	return q.channelID
}

func (q *Queue) AudioStream() *Audio {
	return q.audioStream
}

func (q *Queue) VoiceConnection() *discordgo.VoiceConnection {
	return q.vc
}

func (q *Queue) NowPlaying() *miri.SongResult {
	return q.nowPlaying
}

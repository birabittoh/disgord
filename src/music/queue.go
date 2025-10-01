package music

import (
	"context"

	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

type Queue struct {
	nowPlaying  *miri.SongResult
	items       []miri.SongResult
	audioStream *Audio
	vc          *discordgo.VoiceConnection
	channelID   string
	client      *miri.Client
	ctx         context.Context
}

func (q *Queue) AddTrack(ms *MusicService, track *miri.SongResult) {
	q.AddTracks(ms, []miri.SongResult{*track})
}

func (q *Queue) AddTracks(ms *MusicService, tracks []miri.SongResult) {
	q.items = append(q.items, tracks...)
	if q.nowPlaying == nil {
		err := q.PlayNext(ms, false)
		if err != nil {
			ms.Logger.Error(err)
		}
	}
}

func (q *Queue) PlayNext(ms *MusicService, skip bool) (err error) {
	if q.vc == nil || ms.us.Ctx == nil {
		return
	}

	if q.audioStream != nil && q.audioStream.playing {
		q.audioStream.Stop()
		if skip {
			return nil
		}
	}
	if len(q.items) == 0 {
		ms.DeleteQueue(q.vc.GuildID)
		return nil
	}
	q.nowPlaying = &q.items[0]
	q.items = q.items[1:]
	q.audioStream, err = NewAudio(q.nowPlaying, q.vc, ms)
	if err != nil {
		return
	}
	q.audioStream.Monitor(func() { q.PlayNext(ms, false) })
	return
}

func (q *Queue) Stop() {
	q.Clear()
	if q.audioStream != nil {
		q.audioStream.Stop()
		q.audioStream = nil // Clear the stale audio stream
	}
	q.nowPlaying = nil
	if q.vc != nil && q.ctx != nil {
		q.vc.Disconnect(q.ctx)
	}
	q.vc = nil // Clear the stale connection
}

func (q *Queue) Clear() {
	q.items = []miri.SongResult{}
}

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

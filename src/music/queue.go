package music

import (
	"context"
	"sync"

	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

type Queue struct {
	mu          sync.RWMutex
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
	q.mu.Lock()
	q.items = append(q.items, tracks...)
	isNil := q.nowPlaying == nil
	q.mu.Unlock()

	if isNil {
		err := q.PlayNext(ms, false)
		if err != nil {
			ms.Logger.Error(err)
		}
	}
}

func (q *Queue) PlayNext(ms *MusicService, skip bool) (err error) {
	q.mu.Lock()

	if q.vc == nil || ms.us.Ctx == nil {
		q.mu.Unlock()
		return
	}

	if q.audioStream != nil && q.audioStream.Playing() {
		if skip {
			q.mu.Unlock()
			q.audioStream.Stop()
			return nil
		}
		q.audioStream.Stop()
	}

	if len(q.items) == 0 {
		// ms.DeleteQueue calls q.Stop() which also locks q.mu.
		// To avoid deadlock, we need to unlock q.mu before calling DeleteQueue,
		// but ms.DeleteQueue already locks ms.mu and then q.Stop locks q.mu.
		// Wait, if we are in q.PlayNext, we have q.mu.
		// If we call ms.DeleteQueue, it will try to lock ms.mu.
		// ms.mu should be locked BEFORE q.mu if possible to avoid deadlocks.
		// But here we are already inside a Queue method.

		// Let's release q.mu before calling DeleteQueue.
		guildID := q.vc.GuildID
		q.mu.Unlock()
		ms.DeleteQueue(guildID)
		return nil
	}

	q.nowPlaying = &q.items[0]
	q.items = q.items[1:]
	q.audioStream, err = NewAudio(q.nowPlaying, q.vc, ms, 0)
	if err != nil {
		q.mu.Unlock()
		return
	}

	q.audioStream.SetOnFinish(func() { q.PlayNext(ms, false) })
	q.audioStream.Monitor()
	q.mu.Unlock()
	return
}

func (q *Queue) Seek(ms *MusicService, seekTo int) (err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.vc == nil || ms.us.Ctx == nil {
		return
	}

	if q.audioStream == nil || !q.audioStream.Playing() {
		return
	}

	q.audioStream.SetOnFinish(nil)
	q.audioStream.Stop()

	q.audioStream, err = NewAudio(q.nowPlaying, q.vc, ms, seekTo)
	if err != nil {
		return
	}
	q.audioStream.SetOnFinish(func() { q.PlayNext(ms, false) })
	q.audioStream.Monitor()
	return
}

func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = []miri.SongResult{}
	if q.audioStream != nil {
		q.audioStream.SetOnFinish(nil)
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
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = []miri.SongResult{}
}

func (q *Queue) Tracks() []miri.SongResult {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.nowPlaying != nil {
		return append([]miri.SongResult{*q.nowPlaying}, q.items...)
	}
	return q.items
}

func (q *Queue) VoiceChannelID() string {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.channelID
}

func (q *Queue) AudioStream() *Audio {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.audioStream
}

func (q *Queue) VoiceConnection() *discordgo.VoiceConnection {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.vc
}

func (q *Queue) NowPlaying() *miri.SongResult {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.nowPlaying
}

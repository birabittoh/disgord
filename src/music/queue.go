package music

import (
	"context"
	"sync"

	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

type Queue struct {
	mu          sync.Mutex
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
	defer q.mu.Unlock()

	q.items = append(q.items, tracks...)
	if q.nowPlaying == nil {
		err := q.playNextLocked(ms, false)
		if err != nil {
			ms.Logger.Error(err)
		}
	}
}

func (q *Queue) PlayNext(ms *MusicService, skip bool) (err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.playNextLocked(ms, skip)
}

func (q *Queue) playNextLocked(ms *MusicService, skip bool) (err error) {
	if q.vc == nil || q.ctx == nil {
		return
	}

	if q.audioStream != nil && q.audioStream.IsPlaying() {
		q.audioStream.Stop()
		if skip {
			return nil
		}
	}

	if len(q.items) == 0 {
		q.nowPlaying = nil
		// However, ms.DeleteQueue(q.vc.GuildID) will call q.Stop() which also locks q.mu.
		// To avoid deadlock, we should probably call DeleteQueue in a goroutine or handle it outside.
		// Actually, MusicService.GetQueue already cleans up if q.items is empty.
		go ms.DeleteQueue(q.vc.GuildID)
		return nil
	}

	// Copy the first item and then reslice to avoid pinning the backing array
	track := q.items[0]
	q.nowPlaying = &track
	q.items = q.items[1:]
	q.audioStream, err = NewAudio(q.nowPlaying, q.vc, ms, 0)
	if err != nil {
		return
	}

	q.audioStream.SetOnFinish(func() { q.PlayNext(ms, false) })
	q.audioStream.Monitor()
	return
}

func (q *Queue) Seek(ms *MusicService, seekTo int) (err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.vc == nil || q.ctx == nil {
		return
	}

	if q.audioStream == nil || !q.audioStream.IsPlaying() {
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
		q.audioStream.Stop()
		q.audioStream = nil
	}

	q.nowPlaying = nil
	if q.vc != nil && q.ctx != nil {
		q.vc.Disconnect(q.ctx)
	}
	q.vc = nil
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = []miri.SongResult{}
}

func (q *Queue) Tracks() []miri.SongResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.nowPlaying != nil {
		return append([]miri.SongResult{*q.nowPlaying}, q.items...)
	}
	return q.items
}

func (q *Queue) VoiceChannelID() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.channelID
}

func (q *Queue) AudioStream() *Audio {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.audioStream
}

func (q *Queue) VoiceConnection() *discordgo.VoiceConnection {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.vc
}

func (q *Queue) NowPlaying() *miri.SongResult {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.nowPlaying
}

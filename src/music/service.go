package music

import (
	"os"
	"sync"
	"time"

	"github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/birabittoh/mylo"
	"github.com/bwmarrin/discordgo"
)

type MusicService struct {
	us *globals.UtilsService
	mu sync.RWMutex

	Logger   *mylo.Logger
	Queues   map[string]*Queue
	Searches *globals.LRUCache[string, []miri.SongResult]
}

func NewMusicService(us *globals.UtilsService) (*MusicService, error) {
	return &MusicService{
		us:       us,
		Logger:   mylo.New(os.Stdout, globals.LoggerMusic, us.Config.LogLevel, globals.LogFlags),
		Queues:   make(map[string]*Queue),
		Searches: globals.NewLRUCache[string, []miri.SongResult](100), // Capacity of 100 recent searches
	}, nil
}

func (ms *MusicService) GetVoiceConnection(vc string, guildID string) (voice *discordgo.VoiceConnection, err error) {
	alreadyConnected := false
	for _, vs := range ms.us.Session.VoiceConnections {
		if vs.GuildID == guildID {
			voice = vs
			alreadyConnected = true
			break
		}
	}
	if !alreadyConnected {
		voice, err = ms.us.Session.ChannelVoiceJoin(ms.us.Ctx, guildID, vc, false, true)
		if err != nil {
			ms.Logger.Errorf("could not join voice channel: %v", err)
			return nil, err
		}
	}
	return voice, nil
}

func (ms *MusicService) GetOrCreateQueue(vc *discordgo.VoiceConnection, channelID string) (*Queue, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	q, ok := ms.Queues[vc.GuildID]
	var empty bool
	if ok {
		q.mu.Lock()
		empty = q.nowPlaying == nil && len(q.items) == 0
		q.mu.Unlock()
	}

	if !ok || empty {
		dCfg, err := miri.NewConfig(ms.us.Config.ArlCookie, ms.us.Config.SecretKey)
		if err != nil {
			return nil, err
		}

		dCfg.Timeout = 30 * time.Minute // long timeout for music streaming
		q = &Queue{
			vc:        vc,
			channelID: channelID,
			ctx:       ms.us.Ctx,
		}

		q.client, err = miri.New(ms.us.Ctx, dCfg)
		if err != nil {
			return nil, err
		}
		ms.Queues[vc.GuildID] = q
	} else {
		// Update the voice connection and channel in case they changed
		q.mu.Lock()
		q.vc = vc
		q.channelID = channelID
		q.mu.Unlock()
	}

	return q, nil
}

func (ms *MusicService) GetQueue(guildID string) *Queue {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	q, ok := ms.Queues[guildID]
	if ok {
		q.mu.Lock()
		empty := q.nowPlaying == nil && len(q.items) == 0
		q.mu.Unlock()

		if empty {
			// clean up empty queue
			delete(ms.Queues, guildID)
			return nil
		}
		return q
	}
	return nil
}

func (ms *MusicService) DeleteQueue(guildID string) {
	ms.mu.Lock()
	q, exists := ms.Queues[guildID]
	if !exists {
		ms.mu.Unlock()
		return
	}
	delete(ms.Queues, guildID)
	ms.mu.Unlock()

	ms.Logger.Debugf("Deleting queue for guild %s", guildID)

	q.Stop()
}

type QueueSnapshot struct {
	GuildID        string
	VoiceChannelID string
	Tracks         []miri.SongResult
}

func (ms *MusicService) GetQueuesSnapshot() []QueueSnapshot {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	snapshots := make([]QueueSnapshot, 0, len(ms.Queues))
	for guildID, q := range ms.Queues {
		snapshots = append(snapshots, QueueSnapshot{
			GuildID:        guildID,
			VoiceChannelID: q.VoiceChannelID(),
			Tracks:         q.Tracks(),
		})
	}
	return snapshots
}

func (ms *MusicService) HandleBotVSU(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu.UserID != s.State.User.ID {
		// update is not from this bot
		return
	}

	if vsu.BeforeUpdate == nil {
		// user joined a voice channel
		return
	}

	queue := ms.GetQueue(vsu.GuildID)
	if queue == nil {
		// no queue for this guild
		return
	}

	defer ms.DeleteQueue(vsu.GuildID)

	if queue.NowPlaying() == nil {
		// song has ended naturally
		return
	}

	vc := queue.VoiceConnection()
	if vc == nil {
		return
	}

	if vsu.ChannelID == "" && vsu.BeforeUpdate.ChannelID == queue.VoiceChannelID() {
		ms.Logger.Println("Bot disconnected from voice channel, stopping audio playback.")
		// DeleteQueue will be called in defer
	}
}

package music

import (
	"log/slog"
	"os"
	"time"

	"github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/lmittmann/tint"
)

type PendingSearch struct {
	Results   []miri.SongResult
	MessageID string
}

type MusicService struct {
	us *globals.UtilsService

	Logger   *slog.Logger
	Queues   map[string]*Queue
	Searches *lru.Cache[string, *PendingSearch]
}

func NewMusicService(us *globals.UtilsService) (*MusicService, error) {
	cache, err := lru.New[string, *PendingSearch](128)
	if err != nil {
		return nil, err
	}

	return &MusicService{
		us: us,
		Logger: slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:      us.Config.LogLevel,
			TimeFormat: us.Config.TimeFormat,
		})).With("service", globals.LoggerMusic),
		Queues:   make(map[string]*Queue),
		Searches: cache,
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
			ms.Logger.Error("could not join voice channel", "error", err)
			return nil, err
		}
	}
	return voice, nil
}

func (ms *MusicService) GetOrCreateQueue(vc *discordgo.VoiceConnection, channelID string) (*Queue, error) {
	q := ms.GetQueue(vc.GuildID)
	if q == nil {
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
		q.vc = vc
		q.channelID = channelID
	}

	return q, nil
}

func (ms *MusicService) GetQueue(guildID string) *Queue {
	q, ok := ms.Queues[guildID]
	if ok {
		if q.nowPlaying == nil && len(q.items) == 0 {
			// clean up empty queue
			delete(ms.Queues, guildID)
			return nil
		}
		return q
	}
	return nil
}

func (ms *MusicService) DeleteQueue(guildID string) {
	q, exists := ms.Queues[guildID]
	if !exists {
		return
	}

	ms.Logger.Debug("Deleting queue for guild", "guildID", guildID)

	q.Stop()
	delete(ms.Queues, guildID)
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
		ms.Logger.Info("Bot disconnected from voice channel, stopping audio playback.")
		// DeleteQueue will be called in defer
	}
}

func (ms *MusicService) SetSearchMessageID(channelID, authorID, messageID string) {
	key := getPendingSearchKey(channelID, authorID)
	if ps, ok := ms.Searches.Get(key); ok {
		ps.MessageID = messageID
	}
}

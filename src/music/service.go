package music

import (
	"context"
	"os"
	"time"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/birabittoh/miri"
	"github.com/birabittoh/miri/deezer"
	"github.com/bwmarrin/discordgo"
)

type MusicService struct {
	Ctx    context.Context
	Client *miri.Client
	Logger *mylog.Logger
	Queues map[string]*Queue
}

func NewMusicService(ctx context.Context) (*MusicService, error) {
	cfg, err := deezer.NewConfig(gl.Config.Values.ArlCookie, gl.Config.Values.SecretKey)
	if err != nil {
		return nil, err
	}
	cfg.Timeout = 30 * time.Minute // long timeout for music streaming
	client, err := miri.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	logger := mylog.NewLogger(os.Stdout, "music", mylog.DEBUG)
	return &MusicService{
		Ctx:    ctx,
		Client: client,
		Logger: logger,
		Queues: make(map[string]*Queue),
	}, nil
}

func (ms *MusicService) GetVoiceConnection(vc string, s *discordgo.Session, guildID string) (voice *discordgo.VoiceConnection, err error) {
	alreadyConnected := false
	for _, vs := range s.VoiceConnections {
		if vs.GuildID == guildID {
			voice = vs
			alreadyConnected = true
			break
		}
	}
	if !alreadyConnected {
		voice, err = s.ChannelVoiceJoin(ms.Ctx, guildID, vc, false, true)
		if err != nil {
			ms.Logger.Errorf("could not join voice channel: %v", err)
			return nil, err
		}
	}
	return voice, nil
}

func (ms *MusicService) GetOrCreateQueue(vc *discordgo.VoiceConnection, channelID string) *Queue {
	q := ms.GetQueue(vc.GuildID)
	if q == nil {
		q = &Queue{
			vc:        vc,
			channelID: channelID,
			client:    ms.Client,
			ctx:       ms.Ctx,
		}
		ms.Queues[vc.GuildID] = q
	}

	return q
}

func (ms *MusicService) GetQueue(guildID string) *Queue {
	q, ok := ms.Queues[guildID]
	if ok {
		return q
	}
	return nil
}

func (ms *MusicService) HandleBotVSU(vsu *discordgo.VoiceStateUpdate) {
	if vsu.BeforeUpdate == nil {
		// user joined a voice channel
		return
	}

	queue := ms.GetQueue(vsu.GuildID)
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

	if vsu.ChannelID == "" && vsu.BeforeUpdate.ChannelID == queue.VoiceChannelID() {
		ms.Logger.Println("Bot disconnected from voice channel, stopping audio playback.")
		queue.Stop()
	}
}

package shoot

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"time"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/lmittmann/tint"
)

type ShootService struct {
	us *gl.UtilsService

	logger    *slog.Logger
	magazines *lru.Cache[string, *Magazine]
}

type Magazine struct {
	size uint
	left uint
	last time.Time
}

func NewShootService(us *gl.UtilsService) *ShootService {
	if us.Config.BustProbability == 0 {
		us.Config.BustProbability = 50
	}
	if us.Config.BustProbability > 100 {
		us.Config.BustProbability = 100
	}

	cache, err := lru.New[string, *Magazine](128)
	if err != nil {
		// This should not happen with a hardcoded size of 128
		panic(err)
	}

	return &ShootService{
		us: us,
		logger: slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:      us.Config.LogLevel,
			TimeFormat: time.TimeOnly,
		})).With("service", gl.LoggerShoot),
		magazines: cache,
	}
}

func NewMagazine(size uint) *Magazine {
	return &Magazine{size: size, left: size, last: time.Now()}
}

func (m *Magazine) Update() {
	now := time.Now()
	if m.last.YearDay() != now.YearDay() || m.last.Year() != now.Year() {
		m.left = m.size
	}
}

func (m *Magazine) Left() uint {
	m.Update()
	return m.left
}

func (m Magazine) Size() uint {
	return m.size
}

func (m *Magazine) Shoot() bool {
	if m.Left() <= 0 {
		return false
	}

	m.last = time.Now()
	m.left--
	return true
}

func (m *Magazine) String() string {
	return fmt.Sprintf(gl.MsgMagazineFmt, m.Left(), m.Size())
}

func (ss *ShootService) GetMagazine(userID string) (q *Magazine) {
	q, ok := ss.magazines.Get(userID)
	if ok {
		return
	}

	q = NewMagazine(ss.us.Config.MagazineSize)
	ss.magazines.Add(userID, q)
	return
}

func (ss *ShootService) HandleShoot(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	response, guild, voiceChannelID := ss.us.GetVoiceChannelID(m.Member, m.GuildID, m.Author.ID)
	if voiceChannelID == "" {
		return ss.us.EmbedMessage(response)
	}

	killerID := m.Author.ID
	var allMembers []string
	var err error
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == voiceChannelID && vs.UserID != killerID {
			vs.Member, err = ss.us.Session.State.Member(guild.ID, vs.UserID)
			if err != nil {
				ss.logger.Error("could not get member info", "error", err)
				continue
			}
			if !vs.Member.User.Bot {
				allMembers = append(allMembers, vs.UserID)
			}
		}
	}

	if len(allMembers) == 0 {
		return ss.us.EmbedMessage(fmt.Sprintf(gl.MsgNoOtherUsersFmt, voiceChannelID))
	}

	magazine := ss.GetMagazine(killerID)
	if !magazine.Shoot() {
		return ss.us.EmbedMessage(gl.MsgOutOfBullets)
	}

	victimID := killerID
	if rand.UintN(100) < ss.us.Config.BustProbability {
		victimID = allMembers[rand.IntN(len(allMembers))]
	}

	err = ss.us.Session.GuildMemberMove(m.GuildID, victimID, nil)
	if err != nil {
		ss.logger.Error("could not kick user", "error", err)
		return ss.us.EmbedMessage(gl.MsgCantKickUser)
	}

	return ss.us.EmbedMessage(fmt.Sprintf(gl.MsgShootFmt, victimID, magazine.String()))
}

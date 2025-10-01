package shoot

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

const bustProbability = 50

type ShootService struct {
	logger    *mylog.Logger
	magazines map[string]*Magazine
}

type Magazine struct {
	size uint
	left uint
	last time.Time
}

func NewShootService() *ShootService {
	return &ShootService{
		logger:    mylog.NewLogger(os.Stdin, "shoot", gl.LogLevel),
		magazines: make(map[string]*Magazine),
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

func (s *ShootService) GetMagazine(userID string) (q *Magazine) {
	q, ok := s.magazines[userID]
	if ok {
		return
	}

	q = NewMagazine(gl.Config.Values.MagazineSize)
	s.magazines[userID] = q
	return
}

func (ss *ShootService) HandleShoot(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	response, guild, voiceChannelID := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if voiceChannelID == "" {
		return gl.EmbedMessage(response)
	}

	killerID := m.Author.ID
	var allMembers []string
	var err error
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == voiceChannelID && vs.UserID != killerID {
			vs.Member, err = s.State.Member(guild.ID, vs.UserID)
			if err != nil {
				ss.logger.Errorf("could not get member info: %s", err)
				continue
			}
			if !vs.Member.User.Bot {
				allMembers = append(allMembers, vs.UserID)
			}
		}
	}

	if len(allMembers) == 0 {
		return gl.EmbedMessage(fmt.Sprintf(gl.MsgNoOtherUsersFmt, voiceChannelID))
	}

	magazine := ss.GetMagazine(killerID)
	if !magazine.Shoot() {
		return gl.EmbedMessage(gl.MsgOutOfBullets)
	}

	victimID := killerID
	if rand.IntN(100) < bustProbability {
		victimID = allMembers[rand.IntN(len(allMembers))]
	}

	err = s.GuildMemberMove(m.GuildID, victimID, nil)
	if err != nil {
		ss.logger.Errorf("could not kick user: %s", err)
		return gl.EmbedMessage(gl.MsgCantKickUser)
	}

	return gl.EmbedMessage(fmt.Sprintf(gl.MsgShootFmt, victimID, magazine.String()))
}

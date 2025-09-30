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
	return fmt.Sprintf("_%d/%d bullets left in your magazine._", m.Left(), m.Size())
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
	const bustProbability = 50

	_, err := s.Guild(m.GuildID)
	if err != nil {
		ss.logger.Errorf("could not update guild: %s", err)
		return gl.EmbedMessage(gl.MsgError)
	}

	response, guild, voiceChannelID := gl.GetVoiceChannelID(s, m.Member, m.GuildID, m.Author.ID)
	if voiceChannelID == "" {
		return gl.EmbedMessage(response)
	}

	killerID := m.Author.ID
	var allMembers []string
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == voiceChannelID && vs.UserID != killerID {
			member, err := s.State.Member(guild.ID, vs.UserID)
			if err != nil {
				ss.logger.Errorf("could not get member info: %s", err)
				continue
			}
			if !member.User.Bot {
				allMembers = append(allMembers, vs.UserID)
			}
		}
	}

	if len(allMembers) == 0 {
		return gl.EmbedMessage("There is no one else to shoot in your voice channel.")
	}

	magazine := ss.GetMagazine(killerID)
	if !magazine.Shoot() {
		return gl.EmbedMessage("💨 Too bad... You're out of bullets.")
	}

	var victimID string
	if rand.IntN(100) < bustProbability {
		victimID = killerID
	} else {
		victimID = allMembers[rand.IntN(len(allMembers))]
	}

	err = s.GuildMemberMove(m.GuildID, victimID, nil)
	if err != nil {
		ss.logger.Errorf("could not kick user: %s", err)
		return gl.EmbedMessage("Failed to kick the user from the voice channel.")
	}

	return gl.EmbedMessage("💥 *Bang!* <@" + victimID + "> was shot. " + magazine.String())
}

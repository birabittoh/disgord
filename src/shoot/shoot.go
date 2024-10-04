package shoot

import (
	"fmt"
	"os"
	"time"

	gl "github.com/BiRabittoh/disgord/src/globals"
	"github.com/BiRabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/rand"
)

var logger = mylog.NewLogger(os.Stdin, "shoot", mylog.DEBUG)

type Magazine struct {
	size uint
	left uint
	last time.Time
}

var magazines = map[string]*Magazine{}

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

func GetMagazine(userID string) (q *Magazine) {
	q, ok := magazines[userID]
	if ok {
		return
	}

	q = NewMagazine(gl.Config.Values.MagazineSize)
	magazines[userID] = q
	return
}

func HandleShoot(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	const bustProbability = 50

	_, err := s.Guild(m.GuildID)
	if err != nil {
		logger.Errorf("could not update guild: %s", err)
		return gl.MsgError
	}

	response, guild, voiceChannelID := gl.GetVoiceChannelID(s, m)
	if voiceChannelID == "" {
		return response
	}

	killerID := m.Author.ID
	var allMembers []string
	for _, vs := range guild.VoiceStates {
		logger.Debug(vs.UserID)
		if vs.ChannelID == voiceChannelID && vs.UserID != killerID {
			member, err := s.State.Member(guild.ID, vs.UserID)
			if err != nil {
				logger.Errorf("could not get member info: %s", err)
				continue
			}
			if !member.User.Bot {
				allMembers = append(allMembers, vs.UserID)
			}
		}
	}

	if len(allMembers) == 0 {
		return "There is no one else to shoot in your voice channel."
	}

	magazine := GetMagazine(killerID)
	if !magazine.Shoot() {
		return "💨 Too bad... You're out of bullets."
	}

	var victimID string
	if rand.Intn(100) < bustProbability {
		victimID = killerID
	} else {
		victimID = allMembers[rand.Intn(len(allMembers))]
	}

	err = s.GuildMemberMove(m.GuildID, victimID, nil)
	if err != nil {
		logger.Errorf("could not kick user: %s", err)
		return "Failed to kick the user from the voice channel."
	}

	return "💥 *Bang!* <@" + victimID + "> was shot. " + magazine.String()
}

package globals

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/birabittoh/myks"
	"github.com/bwmarrin/discordgo"
)

var (
	searchPattern = regexp.MustCompile(`watch\?v\x3d([a-zA-Z0-9_-]{11})`)
	searchKS      = myks.New[string](12 * time.Hour)
)

func GetVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) (response string, g *discordgo.Guild, voiceChannelID string) {
	if m.Member == nil {
		response = "Please, use this inside a server."
		return
	}

	g, err := s.State.Guild(m.GuildID)
	if err != nil {
		logger.Errorf("could not get guild: %s", err)
		response = MsgError
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	if voiceChannelID == "" {
		response = "You need to be in a voice channel to use this command."
	}
	return
}

func (bc BotCommand) FormatHelp(command, guildID string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", FormatCommand(bc.ShortCode, guildID))
	}
	return fmt.Sprintf(MsgHelpFmt, FormatCommand(command, guildID)+shortCodeStr, bc.Help)
}

func FormatCommand(command, guildID string) string {
	return fmt.Sprintf("`%s%s`", GetPrefix(guildID), command)
}

func FormatVideo(v string) string {
	return v
}

func ParseUserMessage(messageContent, guildID string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, GetPrefix(guildID))
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, userInput[1:], len(command) > 0
}

func GetPrefix(guildID string) string {
	for _, prefix := range Config.Values.Prefixes {
		if prefix.Name == guildID {
			return prefix.Value
		}
	}

	Config.Values.Prefixes = append(Config.Values.Prefixes, KeyValuePair{Name: guildID, Value: defaultPrefix})
	err := Config.Save()
	if err != nil {
		logger.Errorf("could not save config: %s", err)
	}
	return defaultPrefix
}

func SetPrefix(guildID, prefixValue string) string {
	var found bool
	for i, prefix := range Config.Values.Prefixes {
		if prefix.Name == guildID {
			Config.Values.Prefixes[i].Value = prefixValue
			found = true
			break
		}
	}

	if !found {
		Config.Values.Prefixes = append(Config.Values.Prefixes, KeyValuePair{Name: guildID, Value: prefixValue})
	}

	err := Config.Save()
	if err != nil {
		logger.Errorf("could not save config: %s", err)
	}
	return defaultPrefix
}

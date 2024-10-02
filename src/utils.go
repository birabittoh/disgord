package src

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func getVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) (response string, g *discordgo.Guild, voiceChannelID string) {
	if m.Member == nil {
		response = "Please, use this inside a server."
		return
	}

	_, err := s.Guild(m.GuildID)
	if err != nil {
		logger.Errorf("could not update guild: %s", err)
		response = msgError
		return
	}

	g, err = s.State.Guild(m.GuildID)
	if err != nil {
		logger.Errorf("could not get guild: %s", err)
		response = msgError
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

func formatCommand(command, guildID string) string {
	return fmt.Sprintf("`%s%s`", getPrefix(guildID), command)
}

func parseUserMessage(messageContent, guildID string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, getPrefix(guildID))
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, userInput[1:], len(command) > 0
}

func getPrefix(guildID string) string {
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

func setPrefix(guildID, prefixValue string) string {
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

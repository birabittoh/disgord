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

func formatCommand(command string) string {
	return fmt.Sprintf("`%s%s`", Config.Values.Prefix, command)
}

func parseUserMessage(messageContent string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, Config.Values.Prefix)
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, userInput[1:], len(command) > 0
}

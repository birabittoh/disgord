package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bc BotCommand) FormatHelp(command string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", formatCommand(bc.ShortCode))
	}
	return fmt.Sprintf(helpFmt, formatCommand(command)+shortCodeStr, bc.Help)
}

func getVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) (response string, g *discordgo.Guild, voiceChannelID string) {
	if m.Member == nil {
		response = "Please, use this inside a server."
		return
	}

	g, err := s.State.Guild(m.GuildID)
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
	return fmt.Sprintf("`%s%s`", config.Values.Prefix, command)
}

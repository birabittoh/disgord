package src

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	handlersMap   map[string]BotCommand
	shortCommands = map[string]string{}
)

func (bc BotCommand) FormatHelp(command, guildID string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", formatCommand(bc.ShortCode, guildID))
	}
	return fmt.Sprintf(helpFmt, formatCommand(command, guildID)+shortCodeStr, bc.Help)
}

func InitHandlers() {
	handlersMap = map[string]BotCommand{
		"echo":   {ShortCode: "e", Handler: handleEcho, Help: "echoes a message"},
		"shoot":  {ShortCode: "sh", Handler: handleShoot, Help: "shoots a random user in your voice channel"},
		"prefix": {Handler: handlePrefix, Help: "sets the bot's prefix for this server"},
		"help":   {ShortCode: "h", Handler: handleHelp, Help: "shows this help message"},
	}

	for command, botCommand := range handlersMap {
		if botCommand.ShortCode == "" {
			continue
		}
		shortCommands[botCommand.ShortCode] = command
	}
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) (response string, ok bool, err error) {
	command, args, ok := parseUserMessage(m.Content, m.GuildID)
	if !ok {
		return
	}

	longCommand, short := shortCommands[command]
	if short {
		command = longCommand
	}

	botCommand, found := handlersMap[command]
	if !found {
		response = "Unknown command: " + formatCommand(command, m.GuildID)
		return
	}

	response = botCommand.Handler(args, s, m)
	return
}

func handleEcho(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	return strings.Join(args, " ")
}

func handlePrefix(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	if len(args) == 0 {
		return "Usage: " + formatCommand("prefix <new prefix>", m.GuildID) + "."
	}

	newPrefix := args[0]
	if len(newPrefix) > 10 {
		return "Prefix is too long."
	}

	setPrefix(m.GuildID, newPrefix)

	return "Prefix set to " + newPrefix + "."
}

func handleHelp(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	helpText := "**Bot commands:**\n"
	for command, botCommand := range handlersMap {
		helpText += "* " + botCommand.FormatHelp(command, m.GuildID) + "\n"
	}
	return helpText
}

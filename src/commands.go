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

func (bc BotCommand) FormatHelp(command string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", formatCommand(bc.ShortCode))
	}
	return fmt.Sprintf(helpFmt, formatCommand(command)+shortCodeStr, bc.Help)
}

func InitHandlers() {
	handlersMap = map[string]BotCommand{
		"echo":   {ShortCode: "e", Handler: handleEcho, Help: "echoes a message"},
		"shoot":  {ShortCode: "sh", Handler: handleShoot, Help: "shoots a random user in your voice channel"},
		"prefix": {Handler: handlePrefix, Help: "sets the bot's prefix"},
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
	command, args, ok := parseUserMessage(m.Content)
	if !ok {
		return
	}

	longCommand, short := shortCommands[command]
	if short {
		command = longCommand
	}

	botCommand, found := handlersMap[command]
	if !found {
		response = "Unknown command: " + formatCommand(command)
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
		return "Usage: " + formatCommand("prefix <new prefix>")
	}

	newPrefix := args[0]
	if len(newPrefix) > 10 {
		return "Prefix is too long."
	}

	Config.Values.Prefix = newPrefix
	err := Config.Save()
	if err != nil {
		logger.Errorf("could not save config: %s", err)
	}

	return "Prefix set to " + formatCommand("")
}

func handleHelp(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	helpText := "**Bot commands:**\n"
	for command, botCommand := range handlersMap {
		helpText += "* " + botCommand.FormatHelp(command) + "\n"
	}
	return helpText
}

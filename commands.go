package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	handlersMap   map[string]BotCommand
	shortCommands = map[string]string{}
)

func InitHandlers() {
	handlersMap = map[string]BotCommand{
		"echo":   {ShortCode: "e", Handler: handleEcho, Help: "echoes a message"},
		"prefix": {ShortCode: "pre", Handler: handlePrefix, Help: "sets the bot's prefix"},
		"help":   {ShortCode: "h", Handler: handleHelp, Help: "shows this help message"},
	}

	for command, botCommand := range handlersMap {
		if botCommand.ShortCode == "" {
			continue
		}
		shortCommands[botCommand.ShortCode] = command
	}
}

func handleHelp(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	helpText := "**Bot commands:**\n"
	for command, botCommand := range handlersMap {
		helpText += "* " + botCommand.FormatHelp(command) + "\n"
	}
	return helpText
}

func handleEcho(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	return strings.Join(args, " ")
}

func handlePrefix(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	if len(args) == 0 {
		return "Usage: prefix <new prefix>"
	}

	newPrefix := args[0]
	if len(newPrefix) > 10 {
		return "Prefix is too long"
	}

	config.Values.Prefix = newPrefix
	err := config.Save()
	if err != nil {
		logger.Errorf("could not save config: %s", err)
	}

	return "Prefix set to " + newPrefix
}

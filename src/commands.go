package src

import (
	"fmt"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/shoot"
	"github.com/bwmarrin/discordgo"
)

const (
	MsgUnknownCommand = "Unknown command: %s."
	MsgPrefixSet      = "Prefix set to `%s`."
	MsgPrefixTooLong  = "Prefix is too long."
	MsgUsagePrefix    = "Usage: %s <new prefix>."
	MsgHelp           = "**Bot commands:**\n%s"
	MsgHelpCommandFmt = "* %s\n"
)

var (
	handlersMap   map[string]gl.BotCommand
	shortCommands = map[string]string{}
)

func InitHandlers() {
	handlersMap = map[string]gl.BotCommand{
		"echo":   {ShortCode: "e", Handler: handleEcho, Help: "echoes a message"},
		"shoot":  {ShortCode: "sh", Handler: shoot.HandleShoot, Help: "shoots a random user in your voice channel"},
		"prefix": {Handler: handlePrefix, Help: "sets the bot's prefix for this server"},
		"play":   {ShortCode: "p", Handler: music.HandlePlay, Help: "plays a song from youtube"},
		"pause":  {ShortCode: "pa", Handler: music.HandlePause, Help: "pauses the current song"},
		"resume": {ShortCode: "r", Handler: music.HandleResume, Help: "resumes the current song"},
		"skip":   {ShortCode: "s", Handler: music.HandleSkip, Help: "skips the current song"},
		"queue":  {ShortCode: "q", Handler: music.HandleQueue, Help: "shows the current queue"},
		"clear":  {ShortCode: "c", Handler: music.HandleClear, Help: "clears the current queue"},
		"leave":  {ShortCode: "l", Handler: music.HandleLeave, Help: "leaves the voice channel"},
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
	command, args, ok := gl.ParseUserMessage(m.Content, m.GuildID)
	if !ok {
		return
	}

	longCommand, short := shortCommands[command]
	if short {
		command = longCommand
	}

	botCommand, found := handlersMap[command]
	if !found {
		response = fmt.Sprintf(MsgUnknownCommand, gl.FormatCommand(command, m.GuildID))
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
		return fmt.Sprintf(MsgUsagePrefix, gl.FormatCommand("prefix", m.GuildID))
	}

	newPrefix := args[0]
	if len(newPrefix) > 10 {
		return MsgPrefixTooLong
	}

	gl.SetPrefix(m.GuildID, newPrefix)

	return fmt.Sprintf(MsgPrefixSet, newPrefix)
}

func handleHelp(args []string, s *discordgo.Session, m *discordgo.MessageCreate) string {
	helpText := MsgHelp
	for command, botCommand := range handlersMap {
		// helpText += fmt.Sprintf()
		helpText += fmt.Sprintf(MsgHelpCommandFmt, botCommand.FormatHelp(command, m.GuildID))
	}
	return helpText
}

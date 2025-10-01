package src

import (
	"fmt"
	"os"
	"slices"
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/birabittoh/disgord/src/shoot"
	"github.com/bwmarrin/discordgo"
)

var (
	logger = mylog.NewLogger(os.Stdout, "main", gl.LogLevel)

	cmdMap       map[string]func(arg string, s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageSend
	handlersMap  map[string]gl.BotCommand
	aliasMap     = map[string]string{}
	commandNames []string

	input = []gl.SlashOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "input",
			Description: "command arguments",
			Required:    true,
		},
	}
)

// Exported getter for handlersMap
func HandlersMap() map[string]gl.BotCommand {
	return handlersMap
}

func InitHandlers(ms *music.MusicService, ss *shoot.ShootService) {
	handlersMap = map[string]gl.BotCommand{
		"echo":   {ShortCode: "e", Handler: handleEcho, Help: "echoes a message", SlashOptions: input},
		"prefix": {Handler: handlePrefix, Help: "sets the bot's prefix for this server", SlashOptions: input},
		"play":   {ShortCode: "p", Handler: ms.HandlePlay, Help: "plays a song", SlashOptions: input},
		"search": {ShortCode: "f", Handler: ms.HandleSearch, Help: "searches for a song", SlashOptions: input},
		"lyrics": {ShortCode: "l", Handler: ms.HandleLyrics, Help: "shows the lyrics of the current song"},
		"skip":   {ShortCode: "s", Handler: ms.HandleSkip, Help: "skips the current song"},
		"queue":  {ShortCode: "q", Handler: ms.HandleQueue, Help: "shows the current queue"},
		"clear":  {ShortCode: "c", Handler: ms.HandleClear, Help: "clears the current queue"},
		"leave":  {Alias: "stop", Handler: ms.HandleLeave, Help: "leaves the voice channel"},
		"shoot":  {Alias: "bang", Handler: ss.HandleShoot, Help: "shoots a random user in your voice channel"},
		"help":   {ShortCode: "h", Handler: handleHelp, Help: "shows this help message"},
	}

	for command, botCommand := range handlersMap {
		if botCommand.ShortCode != "" {
			aliasMap[botCommand.ShortCode] = command
		}

		if botCommand.Alias != "" {
			aliasMap[botCommand.Alias] = command
		}
	}

	commandNames = make([]string, 0, len(handlersMap))
	for command := range handlersMap {
		commandNames = append(commandNames, command)
	}

	slices.Sort(commandNames)

	cmdMap = map[string]func(arg string, s *discordgo.Session, i *discordgo.InteractionCreate) *discordgo.MessageSend{
		"choose_track": ms.HandleChooseTrack,
	}
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) (response *discordgo.MessageSend, ok bool, err error) {
	command, args, ok := gl.ParseUserMessage(m.Content, m.GuildID)
	if !ok {
		return
	}

	if aliasTo, isAlias := aliasMap[command]; isAlias {
		command = aliasTo
	}

	botCommand, found := handlersMap[command]
	if !found {
		response = gl.EmbedMessage(fmt.Sprintf(gl.MsgUnknownCommand, gl.FormatCommand(command, m.GuildID)))
		return
	}

	response = botCommand.Handler(args, s, m)
	return
}

func handleEcho(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	return gl.EmbedMessage(strings.Join(args, " "))
}

func handlePrefix(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	var content string
	if len(args) == 0 {
		content = fmt.Sprintf(gl.MsgUsagePrefix, gl.FormatCommand("prefix", m.GuildID))
	} else {
		newPrefix := args[0]
		if len(newPrefix) > 10 {
			content = gl.MsgPrefixTooLong
		} else {
			gl.SetPrefix(m.GuildID, newPrefix)
			content = fmt.Sprintf(gl.MsgPrefixSet, newPrefix)
		}
	}
	return gl.EmbedMessage(content)
}

func handleHelp(args []string, s *discordgo.Session, m *discordgo.MessageCreate) *discordgo.MessageSend {
	helpText := gl.MsgHelp

	for _, command := range commandNames {
		helpText += fmt.Sprintf(gl.MsgUnorderedList, handlersMap[command].FormatHelp(command, m.GuildID))
	}

	return gl.EmbedMessage(helpText)
}

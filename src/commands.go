package src

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/birabittoh/disgord/src/config"
	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/birabittoh/disgord/src/shoot"
	"github.com/birabittoh/disgord/src/ui"
	"github.com/bwmarrin/discordgo"
)

type BotService struct {
	us *gl.UtilsService
	ms *music.MusicService
	ss *shoot.ShootService
	ui *ui.UIService

	logger       *mylog.Logger
	cmdMap       map[string]func(arg string, i *discordgo.InteractionCreate) *discordgo.MessageSend
	handlersMap  map[string]gl.BotCommand
	aliasMap     map[string]string
	commandNames []string
}

func NewBotService(cfg *config.Config) (bs *BotService, err error) {
	bs = &BotService{
		us:       gl.NewUtilsService(cfg),
		aliasMap: make(map[string]string),
	}
	bs.logger = mylog.New(os.Stdout, "main", bs.us.Config.LogLevel)

	bs.us.Session, err = discordgo.New("Bot " + bs.us.Config.Token)
	if err != nil {
		return nil, errors.New("could not create bot session: " + err.Error())
	}

	bs.ss = shoot.NewShootService(bs.us)
	bs.ms, err = music.NewMusicService(bs.us)
	if err != nil {
		return nil, errors.New("could not initialize music service: " + err.Error())
	}
	bs.ui = ui.NewUIService(bs.us, bs.ms)

	bs.initHandlers()
	bs.us.Session.AddHandler(bs.messageHandler)
	bs.us.Session.AddHandler(bs.readyHandler)
	bs.us.Session.AddHandler(bs.slashHandler)
	bs.us.Session.AddHandler(bs.ms.HandleBotVSU)

	return bs, nil
}

func (bs *BotService) Start() error {
	err := bs.us.Session.Open()
	if err != nil {
		return errors.New("could not open session: " + err.Error())
	}

	go func() {
		err := bs.registerSlashCommands()
		if err != nil {
			bs.logger.Errorf("could not register slash commands: %s", err)
		}
	}()

	go bs.ui.Start()

	return nil
}

func (bs *BotService) Stop() {
	if err := bs.us.Session.Close(); err != nil {
		bs.logger.Errorf("could not close session: %s", err)
	}
}

func (bs *BotService) initHandlers() {
	defaultSearchOptions := []gl.SlashOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        gl.DefaultSearchOptionName,
			Description: gl.DefaultSearchOptionDescription,
			Required:    true,
		},
	}

	bs.handlersMap = map[string]gl.BotCommand{
		"echo":   {ShortCode: "e", Handler: bs.handleEcho, Help: "echoes a message", SlashOptions: defaultSearchOptions},
		"play":   {ShortCode: "p", Handler: bs.ms.HandlePlay, Help: "plays a song", SlashOptions: defaultSearchOptions},
		"search": {ShortCode: "f", Handler: bs.ms.HandleSearch, Help: "searches for a song", SlashOptions: defaultSearchOptions},
		"lyrics": {ShortCode: "l", Handler: bs.ms.HandleLyrics, Help: "shows the lyrics of the current song"},
		"skip":   {ShortCode: "s", Handler: bs.ms.HandleSkip, Help: "skips the current song"},
		"queue":  {ShortCode: "q", Handler: bs.ms.HandleQueue, Help: "shows the current queue"},
		"clear":  {ShortCode: "c", Handler: bs.ms.HandleClear, Help: "clears the current queue"},
		"leave":  {Alias: "stop", Handler: bs.ms.HandleLeave, Help: "leaves the voice channel"},
		"shoot":  {Alias: "bang", Handler: bs.ss.HandleShoot, Help: "shoots a random user in your voice channel"},
		"help":   {ShortCode: "h", Handler: bs.handleHelp, Help: "shows this help message"},
	}

	for command, botCommand := range bs.handlersMap {
		if botCommand.ShortCode != "" {
			bs.aliasMap[botCommand.ShortCode] = command
		}

		if botCommand.Alias != "" {
			bs.aliasMap[botCommand.Alias] = command
		}
	}

	bs.commandNames = make([]string, 0, len(bs.handlersMap))
	for command := range bs.handlersMap {
		bs.commandNames = append(bs.commandNames, command)
	}

	slices.Sort(bs.commandNames)

	bs.cmdMap = map[string]func(arg string, i *discordgo.InteractionCreate) *discordgo.MessageSend{
		"choose_track": bs.ms.HandleChooseTrack,
	}
}

func (bs *BotService) getCommand(name string) *gl.BotCommand {
	if aliasTo, isAlias := bs.aliasMap[name]; isAlias {
		name = aliasTo
	}

	botCommand, found := bs.handlersMap[name]
	if !found {
		return nil
	}
	return &botCommand
}

func (bs *BotService) handleCommand(m *discordgo.MessageCreate) (response *discordgo.MessageSend, ok bool, err error) {
	if bs.us.Config.DisablePrefixCommands {
		return nil, false, nil
	}

	command, args, ok := bs.us.ParseUserMessage(m.Content)
	if !ok {
		return
	}

	bc := bs.getCommand(command)
	if bc == nil {
		response = bs.us.EmbedMessage(fmt.Sprintf(gl.MsgUnknownCommand, bs.us.FormatCommand(command)))
		return
	}

	response = bc.Handler(args, m)
	return
}

func (bs *BotService) handleEcho(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	if len(args) == 0 {
		return nil
	}
	return bs.us.EmbedMessage(strings.Join(args, " "))
}

func (bs *BotService) handleHelp(args []string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	helpText := gl.MsgHelp

	for _, command := range bs.commandNames {
		helpText += fmt.Sprintf(gl.MsgUnorderedList, bs.us.FormatHelp(command, bs.handlersMap[command]))
	}

	return bs.us.EmbedMessage(helpText)
}

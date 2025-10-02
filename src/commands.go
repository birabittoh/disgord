package src

import (
	"errors"
	"fmt"
	"os"
	"slices"

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

	logger          *mylog.Logger
	interactionsMap map[string]gl.BotInteraction
	handlersMap     map[string]gl.BotCommand
	aliasMap        map[string]string
	commandNames    []string
}

func NewBotService(cfg *config.Config) (bs *BotService, err error) {
	bs = &BotService{
		us:       gl.NewUtilsService(cfg),
		aliasMap: make(map[string]string),
	}
	bs.logger = mylog.New(os.Stdout, "main", bs.us.Config.LogLevel)

	bs.us.Session, err = discordgo.New("Bot " + bs.us.Config.BotToken)
	if err != nil {
		return nil, errors.New("could not create bot session: " + err.Error())
	}

	if !bs.us.Config.DisableShoot {
		bs.ss = shoot.NewShootService(bs.us)
	}

	if !bs.us.Config.DisableMusic {
		bs.ms, err = music.NewMusicService(bs.us)
		if err != nil {
			return nil, errors.New("could not initialize music service: " + err.Error())
		}
	}

	if !bs.us.Config.DisableUI {
		bs.ui = ui.NewUIService(bs.us, bs.ms)
	}

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

	if bs.ui != nil {
		go bs.ui.Start()
	}

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
		"help":   {ShortCode: "h", Handler: bs.handleHelp, Help: "shows a help message", Tag: "general"},
		"echo":   {ShortCode: "e", Handler: bs.handleEcho, Help: "echoes a message", SlashOptions: defaultSearchOptions, Tag: "general"},
		"play":   {ShortCode: "p", Handler: bs.ms.HandlePlay, Help: "plays a song", SlashOptions: defaultSearchOptions, Tag: "music"},
		"search": {ShortCode: "f", Handler: bs.ms.HandleSearch, Help: "searches for a song", SlashOptions: defaultSearchOptions, Tag: "music"},
		"lyrics": {ShortCode: "l", Handler: bs.ms.HandleLyrics, Help: "shows the lyrics of the current song", Tag: "music"},
		"skip":   {ShortCode: "s", Handler: bs.ms.HandleSkip, Help: "skips the current song", Tag: "music"},
		"queue":  {ShortCode: "q", Handler: bs.ms.HandleQueue, Help: "shows the current queue", Tag: "music"},
		"clear":  {ShortCode: "c", Handler: bs.ms.HandleClear, Help: "clears the current queue", Tag: "music"},
		"leave":  {Alias: "stop", Handler: bs.ms.HandleLeave, Help: "leaves the voice channel", Tag: "music"},
		"shoot":  {Alias: "bang", Handler: bs.ss.HandleShoot, Help: "shoots a random user in your voice channel", Tag: "shoot"},
	}

	bs.interactionsMap = map[string]gl.BotInteraction{
		"choose_track": {Handler: bs.ms.HandleChooseTrack, Tag: "music"},
	}

	for key, cmd := range bs.handlersMap {
		if cmd.Tag == "shoot" && bs.us.Config.DisableShoot {
			delete(bs.handlersMap, key)
		}
		if cmd.Tag == "music" && bs.us.Config.DisableMusic {
			delete(bs.handlersMap, key)
		}
	}

	for key, interaction := range bs.interactionsMap {
		if interaction.Tag == "music" && bs.us.Config.DisableMusic {
			delete(bs.interactionsMap, key)
		}
		if interaction.Tag == "shoot" && bs.us.Config.DisableShoot {
			delete(bs.interactionsMap, key)
		}
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

func (bs *BotService) handleEcho(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	if len(args) == 0 {
		return nil
	}
	return bs.us.EmbedMessage(args)
}

func (bs *BotService) handleHelp(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	helpText := gl.MsgHelp

	for _, command := range bs.commandNames {
		helpText += fmt.Sprintf(gl.MsgUnorderedList, bs.us.FormatHelp(command, bs.handlersMap[command]))
	}

	return bs.us.EmbedMessage(helpText)
}

package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/birabittoh/disgord/src/config"
	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/shoot"
	"github.com/bwmarrin/discordgo"
	"github.com/lmittmann/tint"
)

type BotService struct {
	US *gl.UtilsService
	MS *music.MusicService
	SS *shoot.ShootService

	logger          *slog.Logger
	interactionsMap map[string]gl.BotInteraction
	handlersMap     map[string]gl.BotCommand
	aliasMap        map[string]string
	commandNames    []string
}

func NewBotService(cfg *config.Config) (bs *BotService, err error) {
	bs = &BotService{
		US:       gl.NewUtilsService(cfg),
		aliasMap: make(map[string]string),
	}
	bs.logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      bs.US.Config.LogLevel,
		TimeFormat: cfg.TimeFormat,
	})).With("service", gl.LoggerMain)

	bs.US.Session, err = discordgo.New("Bot " + bs.US.Config.BotToken)
	if err != nil {
		return nil, errors.New("could not create bot session: " + err.Error())
	}

	if !bs.US.Config.DisableShoot {
		bs.SS = shoot.NewShootService(bs.US)
	}

	if !bs.US.Config.DisableMusic {
		bs.MS, err = music.NewMusicService(bs.US)
		if err != nil {
			return nil, errors.New("could not initialize music service: " + err.Error())
		}
	}

	bs.initHandlers()
	bs.US.Session.AddHandler(bs.messageHandler)
	bs.US.Session.AddHandler(bs.readyHandler)
	bs.US.Session.AddHandler(bs.slashHandler)
	bs.US.Session.AddHandler(bs.MS.HandleBotVSU)

	bs.Start()

	return bs, nil
}

func (bs *BotService) Start() error {
	err := bs.US.Session.Open()
	if err != nil {
		return errors.New("could not open session: " + err.Error())
	}

	go func() {
		err := bs.registerSlashCommands()
		if err != nil {
			bs.logger.Error("could not register slash commands", "error", err)
		}
	}()

	bs.logger.Info("Bot started", "commit", gl.CommitID)
	return nil
}

func (bs *BotService) Stop() {
	if err := bs.US.Session.Close(); err != nil {
		bs.logger.Error("could not close session", "error", err)
	}
	bs.logger.Info("Bot stopped")
}

// don't remove the 's' parameter
func (bs *BotService) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.ID == bs.US.Session.State.User.ID {
		return
	}

	bs.logger.Debug("Got a message", "content", m.Content)

	response, ok, err := bs.handleCommand(m)
	if err != nil {
		bs.logger.Error("could not handle command", "error", err)
		return
	}
	if !ok {
		return
	}
	if response != nil {
		_, err := bs.US.Session.ChannelMessageSendComplex(m.ChannelID, response)
		if err != nil {
			bs.logger.Error("could not send message", "error", err)
		}
	}
}

func (bs *BotService) readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: "online",
		AFK:    false,
		Activities: []*discordgo.Activity{
			{
				Name: "FOSS",
				Type: discordgo.ActivityTypeCompeting,
			},
		},
	})
	bs.logger.Info("Logged in", "user", r.User.String())
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
		"play":   {ShortCode: "p", Handler: bs.MS.HandlePlay, Help: "plays a song", SlashOptions: defaultSearchOptions, Tag: "music"},
		"search": {ShortCode: "f", Handler: bs.MS.HandleSearch, Help: "searches for a song", SlashOptions: defaultSearchOptions, Tag: "music"},
		"lyrics": {ShortCode: "l", Handler: bs.MS.HandleLyrics, Help: "shows the lyrics of the current song", Tag: "music"},
		"seek":   {ShortCode: "se", Handler: bs.MS.HandleSeek, Help: "seeks to a specific position in the current song", SlashOptions: defaultSearchOptions, Tag: "music"},
		"skip":   {ShortCode: "s", Handler: bs.MS.HandleSkip, Help: "skips the current song", Tag: "music"},
		"queue":  {ShortCode: "q", Handler: bs.MS.HandleQueue, Help: "shows the current queue", Tag: "music"},
		"clear":  {ShortCode: "c", Handler: bs.MS.HandleClear, Help: "clears the current queue", Tag: "music"},
		"leave":  {Alias: "stop", Handler: bs.MS.HandleLeave, Help: "leaves the voice channel", Tag: "music"},
		"shoot":  {Alias: "bang", Handler: bs.SS.HandleShoot, Help: "shoots a random user in your voice channel", Tag: "shoot"},
	}

	bs.interactionsMap = map[string]gl.BotInteraction{
		"choose_track": {Handler: bs.MS.HandleChooseTrack, Tag: "music"},
	}

	for key, cmd := range bs.handlersMap {
		if cmd.Tag == "shoot" && bs.US.Config.DisableShoot {
			delete(bs.handlersMap, key)
		}
		if cmd.Tag == "music" && bs.US.Config.DisableMusic {
			delete(bs.handlersMap, key)
		}
	}

	for key, interaction := range bs.interactionsMap {
		if interaction.Tag == "music" && bs.US.Config.DisableMusic {
			delete(bs.interactionsMap, key)
		}
		if interaction.Tag == "shoot" && bs.US.Config.DisableShoot {
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
	if bs.US.Config.DisablePrefixCommands {
		return nil, false, nil
	}

	command, args, ok := bs.US.ParseUserMessage(m.Content)
	if !ok {
		return
	}

	bc := bs.getCommand(command)
	if bc == nil {
		response = bs.US.EmbedMessage(fmt.Sprintf(gl.MsgUnknownCommand, bs.US.FormatCommand(command)))
		return
	}

	response = bc.Handler(args, m)
	return
}

func (bs *BotService) handleEcho(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	if len(args) == 0 {
		return nil
	}
	return bs.US.EmbedMessage(args)
}

func (bs *BotService) handleHelp(args string, m *discordgo.MessageCreate) *discordgo.MessageSend {
	helpText := gl.MsgHelp

	for _, command := range bs.commandNames {
		helpText += fmt.Sprintf(gl.MsgUnorderedList, bs.US.FormatHelp(command, bs.handlersMap[command]))
	}

	return bs.US.EmbedMessage(helpText)
}

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/birabittoh/disgord/src"
	g "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

var logger = mylog.NewLogger(os.Stdout, "init", mylog.DEBUG)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		// logger.Debugf("Ignoring own message: %s", m.Content)
		return
	}

	logger.Debug("Got a message: " + m.Content)

	response, ok, err := src.HandleCommand(s, m)
	if err != nil {
		logger.Errorf("could not handle command: %s", err)
		return
	}
	if !ok {
		// not a command
		// not a choose command
		return
	}
	if response != nil {
		_, err := s.ChannelMessageSendComplex(m.ChannelID, response)
		if err != nil {
			logger.Errorf("could not send message: %s", err)
		}
	}
}

func readyHandler(s *discordgo.Session, r *discordgo.Ready) {
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
	logger.Infof("Logged in as %s", r.User.String())
}

func vsuHandler(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu.UserID != s.State.User.ID {
		// update is not from this bot
		return
	}

	music.HandleBotVSU(vsu)
}

func main() {
	logger.Info("Starting bot... Commit " + g.CommitID)
	var err error
	g.Config, err = myconfig.New[g.MyConfig]("config.json")
	if err != nil {
		logger.Errorf("could not load config: %s", err)
	}

	session, err := discordgo.New("Bot " + g.Config.Values.Token)
	if err != nil {
		logger.Fatalf("could not create bot session: %s", err)
	}

	ctx := context.Background()

	music.Init(&ctx)

	src.InitHandlers()
	session.AddHandler(messageHandler)
	session.AddHandler(readyHandler)
	session.AddHandler(vsuHandler)
	src.AddSlashHandler(session)

	err = session.Open()
	if err != nil {
		logger.Fatalf("could not open session: %s", err)
	}

	go func() {
		err := src.RegisterSlashCommands(session)
		if err != nil {
			logger.Errorf("could not register slash commands: %s", err)
		}
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	logger.Info("Stopping gracefully...")
	err = session.Close()
	if err != nil {
		logger.Errorf("could not close session: %s", err)
	}
}

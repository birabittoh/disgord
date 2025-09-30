package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/birabittoh/disgord/src"
	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

var logger = mylog.NewLogger(os.Stdout, "init", gl.LogLevel)

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

func main() {
	logger.Info("Starting bot... Commit " + gl.CommitID)
	var err error
	gl.Config, err = myconfig.New[gl.MyConfig]("config.json")
	if err != nil {
		logger.Errorf("could not load config: %s", err)
	}

	session, err := discordgo.New("Bot " + gl.Config.Values.Token)
	if err != nil {
		logger.Fatalf("could not create bot session: %s", err)
	}

	ctx := context.Background()

	ms, err := music.NewMusicService(ctx)
	if err != nil {
		logger.Fatalf("could not initialize music service: %s", err)
	}

	// Pass ms to handlers as needed (to be updated in src/commands.go, etc.)
	src.InitHandlers(ms)
	session.AddHandler(messageHandler)
	session.AddHandler(readyHandler)
	session.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) { ms.HandleBotVSU(s, vsu) })
	src.AddSlashHandler(session, ms)

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

package main

import (
	"os"
	"os/signal"

	"github.com/BiRabittoh/disgord/src"
	g "github.com/BiRabittoh/disgord/src/globals"
	"github.com/BiRabittoh/disgord/src/myconfig"
	"github.com/BiRabittoh/disgord/src/mylog"
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
		return
	}
	if response == "" {
		logger.Debug("got empty response")
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		logger.Errorf("could not send message: %s", err)
	}
}

func readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	logger.Infof("Logged in as %s", r.User.String())
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

	src.InitHandlers()
	session.AddHandler(messageHandler)
	session.AddHandler(readyHandler)

	err = session.Open()
	if err != nil {
		logger.Fatalf("could not open session: %s", err)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	logger.Info("Stopping gracefully...")
	err = session.Close()
	if err != nil {
		logger.Errorf("could not close session: %s", err)
	}
}

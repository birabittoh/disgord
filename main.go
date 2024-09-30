package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/BiRabittoh/disgord/myconfig"
	"github.com/bwmarrin/discordgo"
)

func parseUserMessage(messageContent string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, config.Values.Prefix)
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, userInput[1:], len(command) > 0
}

func handleCommand(s *discordgo.Session, m *discordgo.MessageCreate) (response string, ok bool, err error) {
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

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger.Debug("got a message: " + m.Content)

	response, ok, err := handleCommand(s, m)
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
	var err error
	config, err = myconfig.New[Config]("config.json")
	if err != nil {
		logger.Errorf("could not load config: %s", err)
	}

	session, err := discordgo.New("Bot " + config.Values.Token)
	if err != nil {
		logger.Fatalf("could not create bot session: %s", err)
	}

	InitHandlers()
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

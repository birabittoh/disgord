package main

import (
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func parseUserMessage(messageContent string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, prefix)
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	if len(userInput) == 0 {
		return
	}

	return userInput[0], userInput[1:], true
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger.Debug("got a message: " + m.Content)
	command, args, ok := parseUserMessage(m.Content)
	if !ok {
		return
	}

	var response string
	switch command {
	case "e", "echo":
		response = strings.Join(args, " ")
	default:
		response = "unknown command: " + command
	}

	_, err := s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		logger.Errorf("could not send message: %s", err)
	}
}

func readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	logger.Infof("Logged in as %s", r.User.String())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Errorf("Error loading .env file: %s", err)
	}

	appID = os.Getenv("APP_ID")
	if appID == "" {
		logger.Fatal("Could not find Discord App ID.")
	}

	discordToken = os.Getenv("AUTH_TOKEN")
	if discordToken == "" {
		logger.Fatal("Could not find Discord bot token.")
	}

	session, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		logger.Fatalf("could not create bot session: %s", err)
	}

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

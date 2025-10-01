package src

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/birabittoh/disgord/src/config"
	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// don't remove the 's' parameter
func (bs *BotService) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == bs.us.Session.State.User.ID {
		return
	}

	bs.logger.Debug("Got a message: " + m.Content)

	response, ok, err := bs.handleCommand(m)
	if err != nil {
		bs.logger.Errorf("could not handle command: %s", err)
		return
	}
	if !ok {
		return
	}
	if response != nil {
		_, err := bs.us.Session.ChannelMessageSendComplex(m.ChannelID, response)
		if err != nil {
			bs.logger.Errorf("could not send message: %s", err)
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
	bs.logger.Infof("Logged in as %s", r.User.String())
}

func getCommitID() string {
	file, err := os.ReadFile(filepath.Join(".git", "HEAD"))
	if err == nil {
		ref := string(file)
		if len(ref) > 5 && ref[:5] == "ref: " {
			refFile, err := os.ReadFile(filepath.Join(".git", ref[5:len(ref)-1]))
			if err == nil {
				return string(refFile[:7])
			}
		} else if len(ref) >= 7 {
			return ref[:7]
		}
	}
	return "unknown"
}

func Main() (err error) {
	godotenv.Load()

	cfg, err := config.New()
	if err != nil {
		return errors.New("could not load config: " + err.Error())
	}

	bs, err := NewBotService(cfg)
	if err != nil {
		return errors.New("could not create bot service: " + err.Error())
	}

	if gl.CommitID == "" {
		gl.CommitID = getCommitID()
	}

	bs.Start()
	bs.logger.Info("Bot started... Commit " + gl.CommitID)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	bs.Stop()

	return nil
}

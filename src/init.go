package src

import (
	"errors"
	"os"
	"os/signal"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/bwmarrin/discordgo"
)

// don't remove the 's' parameter
func (bs *BotService) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == bs.session.State.User.ID {
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
		_, err := bs.session.ChannelMessageSendComplex(m.ChannelID, response)
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

func Main() error {
	var err error
	gl.Config, err = myconfig.New[gl.MyConfig]("config.json")
	if err != nil {
		return errors.New("could not load config: " + err.Error())
	}

	bs, err := NewBotService(gl.Config.Values)
	if err != nil {
		return errors.New("could not create bot service: " + err.Error())
	}

	bs.Start()
	bs.logger.Info("Bot started... Commit " + gl.CommitID)
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	bs.Stop()

	return nil
}

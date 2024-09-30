package main

import (
	"fmt"
	"os"

	"github.com/BiRabittoh/disgord/myconfig"
	"github.com/BiRabittoh/disgord/mylog"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

const (
	helpFmt = "%s - _%s_"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Config struct {
	ApplicationID string `json:"applicationId"`
	Token         string `json:"token"`

	Prefix string `json:"prefix"`

	Outros []KeyValue `json:"outros"`
	Radios []KeyValue `json:"radios"`

	MagazineSize uint `json:"magazineSize"`
}

type BotCommand struct {
	Handler   func([]string, *discordgo.Session, *discordgo.MessageCreate) string
	ShortCode string
	Help      string
}

func formatCommand(command string) string {
	return fmt.Sprintf("`%s%s`", config.Values.Prefix, command)
}

func (bc BotCommand) FormatHelp(command string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", formatCommand(bc.ShortCode))
	}
	return fmt.Sprintf(helpFmt, formatCommand(command)+shortCodeStr, bc.Help)
}

var (
	discordToken string
	appID        string
	config       *myconfig.Config[Config]

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
	yt     = youtube.Client{}
)

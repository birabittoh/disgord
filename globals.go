package main

import (
	"os"

	"github.com/BiRabittoh/disgord/myconfig"
	"github.com/BiRabittoh/disgord/mylog"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

const (
	msgError = "Something went wrong."
	helpFmt  = "%s - _%s_"
)

var (
	config *myconfig.Config[Config]

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
	yt     = youtube.Client{}
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

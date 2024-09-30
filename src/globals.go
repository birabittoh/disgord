package src

import (
	"os"

	"github.com/BiRabittoh/disgord/src/myconfig"
	"github.com/BiRabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

const (
	msgError = "Something went wrong."
	helpFmt  = "%s - _%s_"
)

var (
	CommitID string
	Config   *myconfig.Config[MyConfig]

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
	yt     = youtube.Client{}
)

type KeyValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MyConfig struct {
	ApplicationID string `json:"applicationId"`
	Token         string `json:"token"`

	Prefix string `json:"prefix"`

	Outros []KeyValuePair `json:"outros"`
	Radios []KeyValuePair `json:"radios"`

	MagazineSize uint `json:"magazineSize"`
}

type BotCommand struct {
	Handler   func([]string, *discordgo.Session, *discordgo.MessageCreate) string
	ShortCode string
	Help      string
}

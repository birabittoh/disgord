package globals

import (
	"os"

	"github.com/BiRabittoh/disgord/src/myconfig"
	"github.com/BiRabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

const (
	MsgError      = "Something went wrong."
	helpFmt       = "%s - _%s_"
	defaultPrefix = "$"
)

var (
	CommitID string
	Config   *myconfig.Config[MyConfig]

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
	YT     = youtube.Client{}
)

type KeyValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MyConfig struct {
	ApplicationID string `json:"applicationId"`
	Token         string `json:"token"`

	Prefixes []KeyValuePair `json:"prefixes"`
	Outros   []KeyValuePair `json:"outros"`
	Radios   []KeyValuePair `json:"radios"`

	MagazineSize uint `json:"magazineSize"`
}

type BotCommand struct {
	Handler   func([]string, *discordgo.Session, *discordgo.MessageCreate) string
	ShortCode string
	Help      string
}

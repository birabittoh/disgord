package globals

import (
	"os"

	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

const (
	MsgError      = "Something went wrong."
	MsgHelpFmt    = "%s - _%s_"
	defaultPrefix = "$"
)

var (
	CommitID string
	Config   *myconfig.Config[MyConfig]

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
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

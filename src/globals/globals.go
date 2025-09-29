package globals

import (
	"os"

	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

const (
	MsgError            = "Something went wrong."
	MsgNoResults        = "No results found."
	MsgNoKeywords       = "Please, provide some keywords."
	MsgAddedToQueue     = "Added to queue: %s."
	MsgNothingIsPlaying = "Nothing is playing."
	MsgSameVoiceChannel = "You need to be in the same voice channel to use this command."
	MsgPaused           = "Paused."
	MsgResumed          = "Resumed."
	MsgSkipped          = "Skipped."
	MsgCleared          = "Cleared."
	MsgLeft             = "Left."
	MsgQueueLine        = "%d. %s\n"

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
	ArlCookie     string `json:"arlCookie"`
	SecretKey     string `json:"secretKey"`

	Prefixes []KeyValuePair `json:"prefixes"`
	Radios   []KeyValuePair `json:"radios"`

	MagazineSize uint `json:"magazineSize"`
}

type BotCommand struct {
	Handler   func([]string, *discordgo.Session, *discordgo.MessageCreate) string
	ShortCode string
	Help      string
}

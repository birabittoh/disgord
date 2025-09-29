package globals

import (
	"os"

	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

const (
	MsgError            = "Something went wrong."
	MsgNoResults        = "No results found."
	MsgNoKeywords       = "Please, provide some keywords."
	MsgNothingIsPlaying = "Nothing is playing."
	MsgSameVoiceChannel = "You need to be in the same voice channel to use this command."
	MsgSearchLine       = "%d. %s (%s)\n"
	MsgSearchHelp       = "\nTo pick a song, just type the number of your selection or 0 to cancel.\n"
	MsgChoiceOutOfRange = "Choice out of range. Please pick a number between 1 and %d."
	MsgCanceled         = "Canceled."
	MsgPaused           = "Paused."
	MsgResumed          = "Resumed."
	MsgSkipped          = "Skipped."
	MsgCleared          = "Cleared."
	MsgLeft             = "Left."
	MsgQueueLine        = "%d. %s\n"
	MsgUnknownCommand   = "Unknown command: %s."
	MsgPrefixSet        = "Prefix set to `%s`."
	MsgPrefixTooLong    = "Prefix is too long."
	MsgUsagePrefix      = "Usage: %s <new prefix>."
	MsgHelp             = "**Bot commands:**\n"
	MsgHelpCommandFmt   = "* %s\n"

	MsgHelpFmt     = "%s - _%s_"
	defaultPrefix  = "$"
	AlbumCoverSize = "xl" // "small", "medium", "big", "xl"
)

var (
	CommitID        string
	Config          *myconfig.Config[MyConfig]
	PendingSearches = map[string][]miri.SongResult{}

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

type SlashOption struct {
	Name        string
	Description string
	Type        discordgo.ApplicationCommandOptionType
	Required    bool
}

type BotCommand struct {
	Handler      func([]string, *discordgo.Session, *discordgo.MessageCreate) *discordgo.MessageSend
	ShortCode    string
	Help         string
	SlashOptions []SlashOption // Slash command options
	SlashOnly    bool          // If true, only available as slash command
}

package globals

import (
	"os"

	"github.com/birabittoh/disgord/src/myconfig"
	"github.com/birabittoh/disgord/src/mylog"
	"github.com/bwmarrin/discordgo"
)

const (
	// General messages
	MsgError            = "Something went wrong."
	MsgNoResults        = "No results found."
	MsgNoKeywords       = "Please, provide some keywords."
	MsgNothingIsPlaying = "Nothing is playing."
	MsgUseInServer      = "You can only use this command inside a server."
	MsgSameVoiceChannel = "You need to be in the same voice channel to use this command."
	MsgNoVoiceChannel   = "You need to be in a voice channel to use this command."
	MsgUnknownCommand   = "Unknown command: %s."
	MsgPrefixSet        = "Prefix set to `%s`."
	MsgPrefixTooLong    = "Prefix is too long."
	MsgUsagePrefix      = "Usage: %s <new prefix>."
	MsgHelp             = "**Bot commands:**\n"
	MsgHelpFmt          = "%s - _%s_"
	MsgOrderedList      = "%d. %s\n"
	MsgUnorderedList    = "* %s\n"

	// Shoot messages
	MsgCantKickUser    = "Could not kick user from the voice channel."
	MsgOutOfBullets    = "💨 Too bad... You're out of bullets."
	MsgNoOtherUsersFmt = "There is no one else to shoot in <#%s>."
	MsgMagazineFmt     = "_%d/%d bullets left in your magazine._"
	MsgShootFmt        = "💥 *Bang!* <@%s> was shot. %s"

	// Music messages
	MsgCanceled           = "Canceled."
	MsgPaused             = "Paused."
	MsgResumed            = "Resumed."
	MsgSkipped            = "Skipped."
	MsgCleared            = "Cleared."
	MsgLeft               = "Left."
	MsgNoLyrics           = "No lyrics found for this song."
	MsgInvalidTrackNumber = "Invalid track selection."
	MsgCantFindSearch     = "Could not find your previous search, please try again."

	defaultPrefix                  = "$"
	defaultColor                   = 0xFF73A8
	AlbumCoverSize                 = "xl" // "small", "medium", "big", "xl"
	MaxSearchResults               = 9
	DiscordEmbedDescriptionLimit   = 4096
	DefaultSearchOptionName        = "input"
	DefaultSearchOptionDescription = "command arguments"

	LogLevel = mylog.DEBUG
)

var (
	CommitID string
	Config   *myconfig.Config[MyConfig]

	logger = mylog.NewLogger(os.Stdout, "globals", LogLevel)
)

type KeyValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MyConfig struct {
	ApplicationID string `json:"applicationID"`
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
	Handler      func([]string, *discordgo.MessageCreate) *discordgo.MessageSend
	ShortCode    string
	Alias        string
	Help         string
	SlashOptions []SlashOption // Slash command options
	SlashOnly    bool          // If true, only available as slash command
}

package globals

import (
	"fmt"
	"strings"
	"time"

	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

func GetVoiceChannelID(s *discordgo.Session, member *discordgo.Member, guildID, authorID string) (response string, g *discordgo.Guild, voiceChannelID string) {
	if member == nil {
		response = MsgUseInServer
		return
	}

	g, err := s.State.Guild(guildID)
	if err != nil {
		response = MsgError
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == authorID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	if voiceChannelID == "" {
		response = MsgNoVoiceChannel
	}
	return
}

func (bc BotCommand) FormatHelp(command, guildID string) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", FormatCommand(bc.ShortCode, guildID))
	}
	if bc.Alias != "" {
		shortCodeStr += fmt.Sprintf(" (%s)", FormatCommand(bc.Alias, guildID))
	}
	return fmt.Sprintf(MsgHelpFmt, FormatCommand(command, guildID)+shortCodeStr, bc.Help)
}

func FormatCommand(command, guildID string) string {
	return fmt.Sprintf("`%s%s`", Config.Prefix, command)
}

func FormatTrackLine(v *miri.SongResult) string {
	duration := time.Duration(v.Duration) * time.Second
	return fmt.Sprintf("%s - **%s** (`%s`)", v.Artist.Name, v.Title, duration.String())
}

func ParseUserMessage(messageContent string) (command string, args []string, ok bool) {
	after, found := strings.CutPrefix(messageContent, Config.Prefix)
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, userInput[1:], len(command) > 0
}

func GetPendingSearchKey(channelID, authorID string) string {
	return channelID + ":" + authorID
}

// EmbedMessage returns a MessageSend with a single embed and fixed color.
func EmbedMessage(content string) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Description: content,
				Color:       Config.Color,
			},
		},
	}
}

// EmbedTrackMessage returns a MessageSend with an embed and a cover image.
func EmbedTrackMessage(track *miri.SongResult) *discordgo.MessageSend {
	response := EmbedMessage(fmt.Sprintf("%s\n\n_%s_", track.Artist.Name, track.Album.Title))
	response.Embeds[0].Title = track.Title
	response.Embeds[0].Thumbnail = &discordgo.MessageEmbedThumbnail{URL: track.CoverURL(Config.AlbumCoverSize)}
	return response
}

// EmbedToResponse converts a MessageSend to an InteractionResponse.
func EmbedToResponse(msg *discordgo.MessageSend) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    msg.Content,
			Components: msg.Components,
			Embeds:     msg.Embeds,
		},
	}
}

// InteractionToMessageCreate converts an InteractionCreate to a MessageCreate.
func InteractionToMessageCreate(i *discordgo.InteractionCreate, args []string) *discordgo.MessageCreate {
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			GuildID:   i.GuildID,
			ChannelID: i.ChannelID,
		},
	}
	if i.Member != nil {
		m.Author = i.Member.User
		m.Member = i.Member
	} else if i.User != nil {
		m.Author = i.User
	}

	if len(args) > 0 {
		m.Content = strings.Join(args, " ")
	}

	return m
}

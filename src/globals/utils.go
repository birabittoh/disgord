package globals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/birabittoh/disgord/src/config"
	"github.com/birabittoh/miri"
	"github.com/bwmarrin/discordgo"
)

type UtilsService struct {
	Session *discordgo.Session
	Config  *config.Config
	Ctx     context.Context
}

func NewUtilsService(cfg *config.Config) *UtilsService {
	return &UtilsService{
		Session: nil, // to be set later
		Config:  cfg,
		Ctx:     context.Background(),
	}
}

func (us *UtilsService) GetVoiceChannelID(member *discordgo.Member, guildID, authorID string) (response string, g *discordgo.Guild, voiceChannelID string) {
	if member == nil {
		response = MsgUseInServer
		return
	}

	g, err := us.Session.State.Guild(guildID)
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

func (us *UtilsService) FormatHelp(command string, bc BotCommand) string {
	var shortCodeStr string
	if bc.ShortCode != "" {
		shortCodeStr = fmt.Sprintf(" (%s)", us.FormatCommand(bc.ShortCode))
	}
	if bc.Alias != "" {
		shortCodeStr += fmt.Sprintf(" (%s)", us.FormatCommand(bc.Alias))
	}
	return fmt.Sprintf(MsgHelpFmt, us.FormatCommand(command)+shortCodeStr, bc.Help)
}

func (us *UtilsService) FormatCommand(command string) string {
	return fmt.Sprintf("`%s%s`", us.Config.Prefix, command)
}

func (us *UtilsService) FormatTrackLine(v *miri.SongResult) string {
	duration := time.Duration(v.Duration) * time.Second
	return fmt.Sprintf("%s - **%s** (`%s`)", v.Artist.Name, v.Title, duration.String())
}

func (us *UtilsService) ParseUserMessage(messageContent string) (command string, args string, ok bool) {
	after, found := strings.CutPrefix(messageContent, us.Config.Prefix)
	if !found {
		return
	}

	userInput := strings.Split(after, " ")
	command = strings.ToLower(userInput[0])
	return command, strings.Join(userInput[1:], " "), len(command) > 0
}

// EmbedMessage returns a MessageSend with a single embed and fixed color.
func (us *UtilsService) EmbedMessage(content string) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Description: content,
				Color:       us.Config.Color,
			},
		},
	}
}

// EmbedTrackMessage returns a MessageSend with an embed and a cover image.
func (us *UtilsService) EmbedTrackMessage(track *miri.SongResult) *discordgo.MessageSend {
	response := us.EmbedMessage(fmt.Sprintf("%s\n\n_%s_", track.Artist.Name, track.Album.Title))
	response.Embeds[0].Title = track.Title
	response.Embeds[0].Thumbnail = &discordgo.MessageEmbedThumbnail{URL: track.CoverURL(us.Config.AlbumCoverSize)}
	return response
}

// EmbedToResponse converts a MessageSend to an InteractionResponse.
func (us *UtilsService) EmbedToResponse(msg *discordgo.MessageSend) *discordgo.InteractionResponse {
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
func (us *UtilsService) InteractionToMessageCreate(i *discordgo.InteractionCreate, args string) *discordgo.MessageCreate {
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

	m.Content = args

	return m
}

func (us *UtilsService) GetInviteLink() string {
	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%d&scope=bot", us.Config.ApplicationID, DiscordPermissions)
}

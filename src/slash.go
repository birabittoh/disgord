// Package slash handles Discord slash command registration and interactions.
package src

import (
	"log"

	g "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

// RegisterSlashCommands registers all commands in handlersMap as Discord slash commands.
func RegisterSlashCommands(session *discordgo.Session) error {
	existingCommands, err := session.ApplicationCommands(session.State.User.ID, "")
	if err != nil {
		return err
	}
	for _, cmd := range existingCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, "", cmd.ID)
		if err != nil {
			return err
		}
		log.Printf("Deleted existing command: %s", cmd.Name)
	}
	for name, botCommand := range HandlersMap() {
		// Convert SlashOptions to discordgo.ApplicationCommandOption
		options := []*discordgo.ApplicationCommandOption{}
		for _, opt := range botCommand.SlashOptions {
			options = append(options, &discordgo.ApplicationCommandOption{
				Type:        opt.Type,
				Name:        opt.Name,
				Description: opt.Description,
				Required:    opt.Required,
			})
		}
		cmd := &discordgo.ApplicationCommand{
			Name:        name,
			Description: botCommand.Help,
			Options:     options,
		}
		created, err := session.ApplicationCommandCreate(g.Config.Values.ApplicationID, "", cmd)
		if err != nil {
			return err
		}
		log.Printf("Registered command: %s (ID: %s)", created.Name, created.ID)
	}
	return nil
}

// AddSlashHandler adds a handler for slash command interactions to the session.
func AddSlashHandler(session *discordgo.Session) {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		name := i.ApplicationCommandData().Name
		botCommand, found := HandlersMap()[name]
		if !found {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command.",
				},
			})
			return
		}
		// Extract arguments
		args := []string{}
		for _, opt := range i.ApplicationCommandData().Options {
			if opt.Type == discordgo.ApplicationCommandOptionString {
				args = append(args, opt.StringValue())
			}
		}

		// Create minimal MessageCreate for compatibility
		m := &discordgo.MessageCreate{
			Message: &discordgo.Message{
				GuildID:   i.GuildID,
				ChannelID: i.ChannelID,
				Author:    i.Member.User,
				Member:    i.Member,
			},
		}
		if len(args) > 0 {
			m.Content = args[0]
		}
		response := botCommand.Handler(args, s, m)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	})
}

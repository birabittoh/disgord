// Package slash handles Discord slash command registration and interactions.
package src

import (
	"strings"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

/*
registerSlashCommands efficiently registers all commands in handlersMap as Discord slash commands.
It only deletes obsolete commands, creates new ones, and updates changed ones.
*/
func (bs *BotService) registerSlashCommands() error {
	existingCommands, err := bs.us.Session.ApplicationCommands(bs.us.Session.State.User.ID, "")
	if err != nil {
		return err
	}

	bs.logger.Debug("Slash commands registration started")

	desired := map[string]*discordgo.ApplicationCommand{}
	if !bs.us.Config.DisableSlashCommands {
		for name, botCommand := range bs.HandlersMap() {
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

			desired[name] = cmd

			// Register alias as a separate command if present and non-empty
			if botCommand.Alias != "" {
				aliasCmd := &discordgo.ApplicationCommand{
					Name:        botCommand.Alias,
					Description: botCommand.Help,
					Options:     options,
				}
				desired[botCommand.Alias] = aliasCmd
			}
		}
	}

	// Delete obsolete commands
	for _, cmd := range existingCommands {
		if _, ok := desired[cmd.Name]; !ok {
			err := bs.us.Session.ApplicationCommandDelete(bs.us.Session.State.User.ID, "", cmd.ID)
			if err != nil {
				return err
			}
			bs.logger.Infof("Deleted obsolete command: %s", cmd.Name)
		}
	}

	// Create or update commands
	for name, desiredCmd := range desired {
		var found *discordgo.ApplicationCommand
		for _, cmd := range existingCommands {
			if cmd.Name == name {
				found = cmd
				break
			}
		}
		if found == nil {
			created, err := bs.us.Session.ApplicationCommandCreate(bs.us.Config.ApplicationID, "", desiredCmd)
			if err != nil {
				return err
			}
			bs.logger.Infof("Created new command: %s", created.Name)
		} else {
			// Compare and update if changed
			changed := found.Description != desiredCmd.Description || len(found.Options) != len(desiredCmd.Options)
			if !changed {
				for i, opt := range found.Options {
					dOpt := desiredCmd.Options[i]
					if opt.Name != dOpt.Name || opt.Description != dOpt.Description || opt.Type != dOpt.Type || opt.Required != dOpt.Required {
						changed = true
						break
					}
				}
			}
			if changed {
				updated, err := bs.us.Session.ApplicationCommandEdit(bs.us.Config.ApplicationID, "", found.ID, desiredCmd)
				if err != nil {
					return err
				}
				bs.logger.Infof("Updated command: %s (ID: %s)", updated.Name, updated.ID)
			}
		}
	}

	bs.logger.Info("Slash commands registration completed")
	return nil
}

// slashHandler adds a handler for Discord interactions, routing them to the appropriate command handlers.
func (bs *BotService) slashHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		splitResult := strings.SplitN(customID, ":", 2)
		if len(splitResult) != 2 {
			response := bs.us.EmbedToResponse(bs.us.EmbedMessage(gl.MsgUnknownCommand))
			s.InteractionRespond(i.Interaction, response)
			return
		}

		cmd, arg := splitResult[0], splitResult[1]
		handler, found := bs.cmdMap[cmd]
		if !found {
			response := bs.us.EmbedToResponse(bs.us.EmbedMessage(gl.MsgUnknownCommand))
			s.InteractionRespond(i.Interaction, response)
			return
		}

		response := handler(arg, i)
		if response != nil {
			resp := bs.us.EmbedToResponse(response)
			s.InteractionRespond(i.Interaction, resp)
		}
		return

	case discordgo.InteractionApplicationCommand:
		if bs.us.Config.DisableSlashCommands {
			return
		}

		name := i.ApplicationCommandData().Name
		botCommand, found := bs.HandlersMap()[name]
		if !found {
			response := bs.us.EmbedToResponse(bs.us.EmbedMessage(gl.MsgUnknownCommand))
			s.InteractionRespond(i.Interaction, response)
			return
		}

		args := []string{}
		for _, opt := range i.ApplicationCommandData().Options {
			if opt.Type == discordgo.ApplicationCommandOptionString {
				args = append(args, opt.StringValue())
			}
		}

		m := bs.us.InteractionToMessageCreate(i, args)
		response := bs.us.EmbedToResponse(botCommand.Handler(args, m))
		s.InteractionRespond(i.Interaction, response)

	default:
		bs.logger.Warnf("Unhandled interaction type: %d", i.Type)
		return
	}
}

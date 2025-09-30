// Package slash handles Discord slash command registration and interactions.
package src

import (
	"strconv"
	"strings"

	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/mylog"

	gl "github.com/birabittoh/disgord/src/globals"
	"github.com/bwmarrin/discordgo"
)

var logger = mylog.NewLogger(nil, "main", gl.LogLevel)

/*
RegisterSlashCommands efficiently registers all commands in handlersMap as Discord slash commands.
It only deletes obsolete commands, creates new ones, and updates changed ones.
*/
func RegisterSlashCommands(session *discordgo.Session) error {
	existingCommands, err := session.ApplicationCommands(session.State.User.ID, "")
	if err != nil {
		return err
	}
	desired := map[string]*discordgo.ApplicationCommand{}
	for name, botCommand := range HandlersMap() {
		options := []*discordgo.ApplicationCommandOption{}
		for _, opt := range botCommand.SlashOptions {
			options = append(options, &discordgo.ApplicationCommandOption{
				Type:        opt.Type,
				Name:        opt.Name,
				Description: opt.Description,
				Required:    opt.Required,
			})
		}
		desired[name] = &discordgo.ApplicationCommand{
			Name:        name,
			Description: botCommand.Help,
			Options:     options,
		}
	}

	// Delete obsolete commands
	for _, cmd := range existingCommands {
		if _, ok := desired[cmd.Name]; !ok {
			err := session.ApplicationCommandDelete(session.State.User.ID, "", cmd.ID)
			if err != nil {
				return err
			}
			logger.Infof("Deleted obsolete command: %s", cmd.Name)
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
			created, err := session.ApplicationCommandCreate(gl.Config.Values.ApplicationID, "", desiredCmd)
			if err != nil {
				return err
			}
			logger.Infof("Created new command: %s (ID: %s)", created.Name, created.ID)
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
				updated, err := session.ApplicationCommandEdit(gl.Config.Values.ApplicationID, "", found.ID, desiredCmd)
				if err != nil {
					return err
				}
				logger.Infof("Updated command: %s (ID: %s)", updated.Name, updated.ID)
			}
		}
	}
	return nil
}

// AddSlashHandler adds a handler for slash command interactions to the session.
func AddSlashHandler(session *discordgo.Session, musicService *music.MusicService) {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			customID := i.MessageComponentData().CustomID
			if after, ok := strings.CutPrefix(customID, "choose_track_"); ok {
				trackIdxStr := after
				trackIdx, err := strconv.Atoi(trackIdxStr)
				if err != nil || trackIdx < 0 {
					response := gl.EmbedToResponse(gl.EmbedMessage("Invalid track selection."))
					s.InteractionRespond(i.Interaction, response)
					return
				}

				key := gl.GetPendingSearchKey(i.ChannelID, i.Member.User.ID)
				results, found := musicService.Searches[key]
				if !found || trackIdx > len(results) {
					response := gl.EmbedToResponse(gl.EmbedMessage("Track not found."))
					s.InteractionRespond(i.Interaction, response)
					return
				}

				if trackIdx == 0 {
					// Cancel selection
					delete(musicService.Searches, key)
					s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
					return
				}

				track := &results[trackIdx-1]
				r, _, vc := gl.GetVoiceChannelID(s, i.Member, i.GuildID, i.Member.User.ID)
				if r != "" {
					response := gl.EmbedToResponse(gl.EmbedMessage(r))
					s.InteractionRespond(i.Interaction, response)
					return
				}

				voice, err := musicService.GetVoiceConnection(vc, s, i.GuildID)
				if err != nil {
					response := gl.EmbedToResponse(gl.EmbedMessage(err.Error()))
					s.InteractionRespond(i.Interaction, response)
					return
				}

				q := musicService.GetOrCreateQueue(voice, vc)
				q.AddTrack(musicService, track)
				delete(musicService.Searches, key)
				defer s.ChannelMessageDelete(i.ChannelID, i.Message.ID)

				coverURL := track.CoverURL(gl.AlbumCoverSize)
				response := gl.EmbedToResponse(gl.EmbedTrackMessage(gl.FormatTrack(track), coverURL))
				err = s.InteractionRespond(i.Interaction, response)
				if err != nil {
					logger.Errorf("could not respond to interaction: %s", err)
				}
				return
			}
		}
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		name := i.ApplicationCommandData().Name
		botCommand, found := HandlersMap()[name]
		if !found {
			response := gl.EmbedToResponse(gl.EmbedMessage("Unknown command."))
			s.InteractionRespond(i.Interaction, response)
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
		response := gl.EmbedToResponse(botCommand.Handler(args, s, m))
		s.InteractionRespond(i.Interaction, response)
	})
}

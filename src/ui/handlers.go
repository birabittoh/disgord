package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/birabittoh/disgord/src/globals"
)

func (ui *UIService) indexHandler(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	err := ui.indexTemplate.Execute(&b, map[string]any{
		"BotName":  ui.us.Session.State.User.Username,
		"CommitID": globals.CommitID,
		"Link":     ui.us.GetInviteLink(),
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(b.Bytes())
}

func (ui *UIService) guildsHandler(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, ui.us.Session.State.Guilds)
}

func (ui *UIService) queuesHandler(w http.ResponseWriter, r *http.Request) {
	if ui.ms == nil {
		jsonSuccess(w, []any{})
		return
	}

	response := []map[string]any{}
	for guildID, queue := range ui.ms.Queues {
		response = append(response, map[string]any{
			"guild_id":   guildID,
			"channel_id": queue.VoiceChannelID(),
			"tracks":     queue.Tracks(), // first track is currently playing
		})
	}
	jsonSuccess(w, response)
}

func (ui *UIService) queuesCommandsHandler(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, ui.queueCmds)
}

func (ui *UIService) queuesCommandHandler(w http.ResponseWriter, r *http.Request) {
	if ui.ms == nil {
		jsonError(w, "Music service is disabled", http.StatusServiceUnavailable)
		return
	}

	guildID := r.PathValue("guild_id")
	if guildID == "" {
		jsonError(w, "Guild ID is required", http.StatusBadRequest)
		return
	}

	var payload QueueCommandPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		jsonError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	command, ok := ui.validQueueCmds[payload.Command]
	if !ok {
		jsonError(w, "Invalid command", http.StatusBadRequest)
		return
	}

	err = command(guildID, payload)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ui *UIService) guildLeaveHandler(w http.ResponseWriter, r *http.Request) {
	guildID := r.PathValue("id")
	if guildID == "" {
		jsonError(w, "Guild ID is required", http.StatusBadRequest)
		return
	}

	err := ui.us.Session.GuildLeave(guildID)
	if err != nil {
		jsonError(w, "Failed to leave guild", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Command Handlers

func (ui *UIService) handleQueuePlay(guildID string, payload QueueCommandPayload) error {
	if payload.VoiceChannelID == "" {
		return errors.New("VoiceChannelID is required for play command")
	}

	_, _, err := ui.ms.PlayToVC(payload.Args, payload.VoiceChannelID, guildID)
	return err
}

func (ui *UIService) handleQueueClear(guildID string, payload QueueCommandPayload) error {
	queue := ui.ms.GetQueue(guildID)
	if queue == nil {
		return errors.New("no active queue for this guild")
	}
	queue.Clear()
	return nil
}

func (ui *UIService) handleQueueSkip(guildID string, payload QueueCommandPayload) error {
	queue := ui.ms.GetQueue(guildID)
	if queue == nil {
		return errors.New("no active queue for this guild")
	}
	return queue.PlayNext(ui.ms, true)
}

func (ui *UIService) handleQueueStop(guildID string, payload QueueCommandPayload) error {
	ui.ms.DeleteQueue(guildID)
	return nil
}

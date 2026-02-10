package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/birabittoh/disgord/src/bot"
	"github.com/birabittoh/disgord/src/globals"
)

func (ui *UIService) indexHandler(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	err := ui.indexTemplate.Execute(&b, map[string]any{
		"botName":    ui.botName,
		"inviteLink": ui.inviteLink,
		"commitID":   globals.CommitID,
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
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	if ui.bs == nil {
		jsonSuccess(w, []any{})
		return
	}

	jsonSuccess(w, ui.bs.US.Session.State.Guilds)
}

func (ui *UIService) queuesHandler(w http.ResponseWriter, r *http.Request) {
	if !ui.IsBotEnabled() || ui.bs.MS == nil {
		jsonSuccess(w, []any{})
		return
	}

	snapshots := ui.bs.MS.GetQueuesSnapshot()
	response := make([]map[string]any, 0, len(snapshots))
	for _, q := range snapshots {
		response = append(response, map[string]any{
			"guild_id":   q.GuildID,
			"channel_id": q.VoiceChannelID,
			"tracks":     q.Tracks,
		})
	}
	jsonSuccess(w, response)
}

func (ui *UIService) queuesCommandsHandler(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, ui.queueCmds)
}

func (ui *UIService) queuesCommandHandler(w http.ResponseWriter, r *http.Request) {
	if ui.bs.MS == nil {
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

	ui.mu.RLock()
	defer ui.mu.RUnlock()
	if ui.bs == nil {
		jsonError(w, "Bot is disabled", http.StatusServiceUnavailable)
		return
	}

	err := ui.bs.US.Session.GuildLeave(guildID)
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

	ui.mu.RLock()
	defer ui.mu.RUnlock()
	if ui.bs == nil || ui.bs.MS == nil {
		return errors.New("bot or music service is disabled")
	}

	_, _, err := ui.bs.MS.PlayToVC(payload.Args, payload.VoiceChannelID, guildID)
	return err
}

func (ui *UIService) handleQueueClear(guildID string, payload QueueCommandPayload) error {
	ui.mu.RLock()
	defer ui.mu.RUnlock()
	if ui.bs == nil || ui.bs.MS == nil {
		return errors.New("bot or music service is disabled")
	}

	queue := ui.bs.MS.GetQueue(guildID)
	if queue == nil {
		return errors.New("no active queue for this guild")
	}
	queue.Clear()
	return nil
}

func (ui *UIService) handleQueueSkip(guildID string, payload QueueCommandPayload) error {
	ui.mu.RLock()
	defer ui.mu.RUnlock()
	if ui.bs == nil || ui.bs.MS == nil {
		return errors.New("bot or music service is disabled")
	}

	queue := ui.bs.MS.GetQueue(guildID)
	if queue == nil {
		return errors.New("no active queue for this guild")
	}
	return queue.PlayNext(ui.bs.MS, true)
}

func (ui *UIService) handleQueueStop(guildID string, payload QueueCommandPayload) error {
	ui.mu.RLock()
	defer ui.mu.RUnlock()
	if ui.bs == nil || ui.bs.MS == nil {
		return errors.New("bot or music service is disabled")
	}

	ui.bs.MS.DeleteQueue(guildID)
	return nil
}

func (ui *UIService) getBotStateHandler(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, EnabledPayload{Enabled: ui.IsBotEnabled()})
}

func (ui *UIService) postBotStateHandler(w http.ResponseWriter, r *http.Request) {
	var payload EnabledPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ui.mu.Lock()
	if payload.Enabled {
		if ui.bs == nil {
			var err error
			ui.bs, err = bot.NewBotService(ui.cfg)
			if err != nil {
				ui.mu.Unlock()
				jsonError(w, "Failed to create bot service: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		if ui.bs != nil {
			ui.bs.Stop()
			ui.bs = nil
		}
	}
	ui.mu.Unlock()

	jsonSuccess(w, EnabledPayload{Enabled: ui.IsBotEnabled()})
}

func (ui *UIService) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

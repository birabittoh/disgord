package ui

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"

	"github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/disgord/src/music"
	"github.com/birabittoh/disgord/src/mylog"
)

const sep = string(os.PathSeparator)

type UIService struct {
	us     *globals.UtilsService
	ms     *music.MusicService
	logger *mylog.Logger

	mux            *http.ServeMux
	indexTemplate  *template.Template
	validQueueCmds map[string]func(string, QueueCommandPayload) error
	queueCmds      []string
}

type QueueCommandPayload struct {
	Command        string `json:"command"`
	Args           string `json:"args,omitempty"`
	VoiceChannelID string `json:"voice_channel_id,omitempty"`
}

func NewUIService(us *globals.UtilsService, ms *music.MusicService) *UIService {
	ui := &UIService{
		us:            us,
		ms:            ms,
		logger:        mylog.New(os.Stdout, "ui", us.Config.LogLevel),
		mux:           http.NewServeMux(),
		indexTemplate: template.Must(template.ParseFiles("templates" + sep + "index.html")),
	}

	ui.validQueueCmds = map[string]func(string, QueueCommandPayload) error{
		"play":  ui.handleQueuePlay, // requires VoiceChannelID
		"clear": ui.handleQueueClear,
		"skip":  ui.handleQueueSkip,
		"stop":  ui.handleQueueStop,
	}

	ui.mux.HandleFunc("GET /", ui.indexHandler)
	ui.mux.HandleFunc("GET /api/guilds", ui.guildsHandler)
	ui.mux.HandleFunc("GET /api/queues", ui.queuesHandler)
	ui.mux.HandleFunc("POST /api/guilds/{id}/leave", ui.guildLeaveHandler)
	ui.mux.HandleFunc("GET /api/queues/commands", ui.queuesCommandsHandler)
	ui.mux.HandleFunc("POST /api/queues/{guild_id}", ui.queuesCommandHandler)
	ui.mux.HandleFunc("GET /healthz", ui.healthzHandler)

	return ui
}

func (ui *UIService) Start() error {
	ui.logger.Infof("Starting UI server on %s", ui.us.Config.UIAddress)
	return http.ListenAndServe(ui.us.Config.UIAddress, ui.mux)
}

// Helper functions
func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, map[string]string{"error": message}, status)
}

func jsonSuccess(w http.ResponseWriter, data any) {
	jsonResponse(w, data, http.StatusOK)
}

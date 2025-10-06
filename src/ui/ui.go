package ui

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"os/signal"

	"github.com/birabittoh/disgord/src/bot"
	"github.com/birabittoh/disgord/src/globals"
	"github.com/birabittoh/mylo"
)

type UIService struct {
	bs     *bot.BotService
	us     *globals.UtilsService
	sigch  chan os.Signal
	logger *mylo.Logger

	mux            *http.ServeMux
	indexTemplate  *template.Template
	validQueueCmds map[string]func(string, QueueCommandPayload) error
	queueCmds      []string
	botName        string
	inviteLink     string
}

type QueueCommandPayload struct {
	Command        string `json:"command"`
	Args           string `json:"args,omitempty"`
	VoiceChannelID string `json:"voice_channel_id,omitempty"`
}

type EnabledPayload struct {
	Enabled bool `json:"enabled"`
}

func NewUIService(bs *bot.BotService) *UIService {
	ui := &UIService{
		bs:            bs,
		us:            bs.US,
		logger:        mylo.New(os.Stdout, globals.LoggerUI, bs.US.Config.LogLevel, globals.LogFlags),
		mux:           http.NewServeMux(),
		indexTemplate: template.Must(template.ParseFiles("templates" + globals.Sep + "index.html")),
		botName:       bs.US.Session.State.User.Username,
		inviteLink:    bs.US.GetInviteLink(),
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
	ui.mux.HandleFunc("GET /api/bot/state", ui.getBotStateHandler)
	ui.mux.HandleFunc("POST /api/bot/state", ui.postBotStateHandler)
	ui.mux.HandleFunc("GET /healthz", ui.healthzHandler)

	return ui
}

func (ui *UIService) Start() error {
	ui.logger.Infof("Starting UI server on %s", ui.us.Config.UIAddress)
	go func() {
		if err := http.ListenAndServe(ui.us.Config.UIAddress, ui.mux); err != nil {
			ui.logger.Errorf("UI server error: %s", err)
			os.Exit(1)
		}
	}()

	ui.sigch = make(chan os.Signal, 1)
	signal.Notify(ui.sigch, os.Interrupt)
	<-ui.sigch

	ui.bs.Stop()
	return nil
}

func (ui *UIService) IsBotEnabled() bool {
	return ui.bs != nil
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

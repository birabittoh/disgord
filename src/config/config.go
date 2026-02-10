package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// General settings
	ApplicationID string // required
	BotToken      string // required
	LogLevel      slog.Level
	TimeFormat    string
	Prefix        string
	Color         int
	UIAddress     string

	// Music settings
	ArlCookie        string // required for music
	SecretKey        string // requred for music
	AlbumCoverSize   string
	MaxSearchResults uint64

	// Shoot settings
	MagazineSize    uint
	BustProbability uint

	// Disable settings
	DisableUI             bool
	DisableSlashCommands  bool
	DisablePrefixCommands bool
	DisableMusic          bool
	DisableShoot          bool
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	panic(errors.New("environment variable " + key + " is required"))
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvUint(key string, fallback uint) uint {
	if value, exists := os.LookupEnv(key); exists {
		res, err := strconv.Atoi(value)
		if err == nil {
			return uint(res)
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		res, err := strconv.ParseBool(value)
		if err == nil {
			return res
		}
	}
	return fallback
}

func New() (*Config, error) {
	colorStr := getEnv("COLOR", "FF73A8")
	color, err := strconv.ParseInt(colorStr, 16, 32)
	if err != nil {
		color = 0xFF73A8
	}

	c := &Config{
		ApplicationID: requireEnv("APPLICATION_ID"),
		BotToken:      requireEnv("BOT_TOKEN"),
		LogLevel:      slog.LevelInfo,
		Prefix:        getEnv("PREFIX", "$"),
		Color:         int(color),
		UIAddress:     getEnv("UI_ADDRESS", ":8080"),

		ArlCookie:        getEnv("ARL_COOKIE", ""),
		SecretKey:        getEnv("SECRET_KEY", ""),
		AlbumCoverSize:   getEnv("ALBUM_COVER_SIZE", "xl"),
		MaxSearchResults: uint64(getEnvUint("MAX_SEARCH_RESULTS", 9)),

		MagazineSize:    getEnvUint("MAGAZINE_SIZE", 3),
		BustProbability: getEnvUint("BUST_PROBABILITY", 50),

		DisableUI:             getEnvBool("DISABLE_UI", false),
		DisableSlashCommands:  getEnvBool("DISABLE_SLASH_COMMANDS", false),
		DisablePrefixCommands: getEnvBool("DISABLE_PREFIX_COMMANDS", false),
		DisableMusic:          getEnvBool("DISABLE_MUSIC", false),
		DisableShoot:          getEnvBool("DISABLE_SHOOT", false),
		TimeFormat:            time.RFC3339,
	}

	if getEnvBool("DEBUG", false) {
		c.LogLevel = slog.LevelDebug
	}

	return c, c.Validate()
}

func (c *Config) Validate() error {
	if c.ApplicationID == "" || c.BotToken == "" {
		return errors.New("application id and bot token must be set")
	}

	l := len(c.Prefix)
	if l == 0 || l > 5 {
		return errors.New("prefix must be between 1 and 5 characters long")
	}

	if c.Color < 0 || c.Color > 0xFFFFFF {
		return errors.New("color must be a valid hex color code")
	}

	if c.UIAddress == "" {
		return errors.New("UI address must be set")
	}

	if !c.DisableMusic && (c.ArlCookie == "" || c.SecretKey == "") {
		return errors.New("ARL_COOKIE and SECRET_KEY must be set if DISABLE_MUSIC is false")
	}

	allowedSizes := map[string]bool{
		"small":  true,
		"medium": true,
		"big":    true,
		"xl":     true,
	}
	if !allowedSizes[c.AlbumCoverSize] {
		return errors.New("album cover size must be one of: small, medium, big, xl")
	}

	if c.MaxSearchResults == 0 || c.MaxSearchResults > 100 {
		return errors.New("max search results must be between 1 and 100")
	}

	if c.BustProbability > 100 {
		return errors.New("bust probability must be between 0 and 100")
	}

	if c.DisableSlashCommands && c.DisablePrefixCommands {
		return errors.New("at least one of slash commands or prefix commands must be enabled")
	}

	return nil
}

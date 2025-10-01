package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/birabittoh/disgord/src/mylog"
)

type Config struct {
	ApplicationID string // required
	Token         string // required
	ArlCookie     string // required
	SecretKey     string // required

	LogLevel              mylog.Level
	Prefix                string
	Color                 int
	DisableSlashCommands  bool
	DisablePrefixCommands bool

	AlbumCoverSize   string
	MaxSearchResults uint64

	MagazineSize    uint
	BustProbability uint
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
		// Required fields
		ApplicationID: requireEnv("APPLICATION_ID"),
		Token:         requireEnv("BOT_TOKEN"),
		ArlCookie:     requireEnv("ARL_COOKIE"),
		SecretKey:     requireEnv("SECRET_KEY"),

		// General settings
		LogLevel:              mylog.INFO,
		Prefix:                getEnv("PREFIX", "$"),
		Color:                 int(color),
		DisableSlashCommands:  getEnvBool("DISABLE_SLASH_COMMANDS", false),
		DisablePrefixCommands: getEnvBool("DISABLE_PREFIX_COMMANDS", false),

		// Music settings
		AlbumCoverSize:   getEnv("ALBUM_COVER_SIZE", "xl"),
		MaxSearchResults: uint64(getEnvUint("MAX_SEARCH_RESULTS", 9)),

		// Shoot settings
		MagazineSize:    getEnvUint("MAGAZINE_SIZE", 3),
		BustProbability: getEnvUint("BUST_PROBABILITY", 50),
	}

	if getEnvBool("DEBUG", false) {
		c.LogLevel = mylog.DEBUG
	}

	return c, c.Validate()
}

func (c *Config) Validate() error {
	if c.ApplicationID == "" || c.Token == "" || c.ArlCookie == "" || c.SecretKey == "" {
		return errors.New("all required fields must be set")
	}

	l := len(c.Prefix)
	if l == 0 || l > 5 {
		return errors.New("prefix must be between 1 and 5 characters long")
	}

	if c.Color < 0 || c.Color > 0xFFFFFF {
		return errors.New("color must be a valid hex color code")
	}

	if c.DisableSlashCommands && c.DisablePrefixCommands {
		return errors.New("at least one of slash commands or prefix commands must be enabled")
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

	return nil
}

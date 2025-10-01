package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	ApplicationID string // required
	Token         string // required
	ArlCookie     string // required
	SecretKey     string // required

	Prefix          string
	Color           int
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

func New() (*Config, error) {
	colorStr := getEnv("COLOR", "FF73A8")
	color, err := strconv.ParseInt(colorStr, 16, 32)
	if err != nil {
		color = 0xFF73A8
	}

	c := &Config{
		ApplicationID: requireEnv("APPLICATION_ID"),
		Token:         requireEnv("BOT_TOKEN"),
		ArlCookie:     requireEnv("ARL_COOKIE"),
		SecretKey:     requireEnv("SECRET_KEY"),

		Prefix:          getEnv("CMD_PREFIX", "$"),
		Color:           int(color),
		MagazineSize:    getEnvUint("MAGAZINE_SIZE", 3),
		BustProbability: getEnvUint("BUST_PROBABILITY", 50),
	}

	return c, c.Validate()
}

func (c *Config) Validate() error {
	if c.ApplicationID == "" || c.Token == "" || c.ArlCookie == "" || c.SecretKey == "" {
		return errors.New("all required fields must be set")
	}
	if c.BustProbability > 100 {
		return errors.New("bust probability must be between 0 and 100")
	}
	if c.MagazineSize == 0 {
		return errors.New("magazine size must be greater than 0")
	}

	if len(c.Prefix) == 0 {
		return errors.New("command prefix cannot be empty")
	}
	return nil
}

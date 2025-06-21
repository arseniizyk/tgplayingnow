package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config interface {
	SpotifyClientID() string
	SpotifyClientSecret() string
	TelegramAppId() int32
	TelegramAppHash() string
}

type envConfig struct {
	spotifyAppId     string
	spotifyAppSecret string
	telegramAppId    int32
	telegramAppHash  string
}

var (
	ErrEmptyEnv  = errors.New("env variable is empty")
	ErrInvalidWD = errors.New("can't get working directory")
)

func New() (Config, error) {
	cfg, err := loadConfig()
	if err != nil || cfg.validate() != nil {
		return nil, err
	}

	return cfg, nil
}

func loadConfig() (*envConfig, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, ErrInvalidWD
	}

	path := filepath.Join(pwd, "../.env")

	if err := godotenv.Load(path); err != nil {
		return nil, err
	}

	return &envConfig{
		spotifyAppId:     os.Getenv("SPOTIFY_CLIENT_ID"),
		spotifyAppSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		telegramAppId:    getEnvInt32("TELEGRAM_APP_ID"),
		telegramAppHash:  os.Getenv("TELEGRAM_APP_HASH"),
	}, nil
}

func (c *envConfig) SpotifyClientID() string     { return c.spotifyAppId }
func (c *envConfig) SpotifyClientSecret() string { return c.spotifyAppSecret }
func (c *envConfig) TelegramAppId() int32        { return c.telegramAppId }
func (c *envConfig) TelegramAppHash() string     { return c.telegramAppHash }

func (c *envConfig) validate() error {
	if c.spotifyAppId == "" || c.spotifyAppSecret == "" ||
		c.telegramAppHash == "" || c.telegramAppId == 0 {
		return ErrEmptyEnv
	}

	return nil
}

func getEnvInt32(key string) int32 {
	val := os.Getenv(key)

	num, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		log.Println("cannot convert telegram id to int32")
		panic(err)
	}

	return int32(num)
}

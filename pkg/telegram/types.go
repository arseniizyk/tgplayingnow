package telegram

import (
	"path/filepath"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/zelenin/go-tdlib/client"
)

type Telegram struct {
	cfg    config.Config
	oldBio string
	c      *client.Client
}

func New(cfg config.Config) *Telegram {
	return &Telegram{
		cfg: cfg,
	}
}

func generateParams(appId int32, appHash string) *client.SetTdlibParametersRequest {
	client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{ //nolint:all
		NewVerbosityLevel: 0,
	})

	return &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     false,
		UseChatInfoDatabase: false,
		UseMessageDatabase:  false,
		UseSecretChats:      false,
		ApiId:               appId,
		ApiHash:             appHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}
}

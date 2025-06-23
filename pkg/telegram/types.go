package telegram

import (
	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/zelenin/go-tdlib/client"
)

type telegram struct {
	cfg    config.Config
	OldBio string
	client *client.Client
}

func New(cfg config.Config) Telegram {
	return &telegram{
		cfg: cfg,
	}
}

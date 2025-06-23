package main

import (
	"log"

	"github.com/arseniizyk/tgplayingnow/internal/app"
	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/spotify"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
	"github.com/arseniizyk/tgplayingnow/pkg/telegram"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.New()
	if err != nil {
		log.Fatal(err)
	}

	spotify := spotify.New(cfg, storage)
	telegram := telegram.New(cfg)

	app := app.New(cfg, storage, spotify, telegram)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

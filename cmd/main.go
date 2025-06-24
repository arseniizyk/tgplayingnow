package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arseniizyk/tgplayingnow/internal/app"
	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/spotify"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
	"github.com/arseniizyk/tgplayingnow/pkg/telegram"
)

func main() {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

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

	go func() {
		if err := app.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	<-sigCh
	if err := telegram.ResetBio(); err != nil {
		log.Fatalf("failed to reset bio: %v", err)
	}
}

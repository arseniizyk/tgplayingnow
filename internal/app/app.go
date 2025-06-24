package app

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/spotify"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
	"github.com/arseniizyk/tgplayingnow/pkg/telegram"
)

type App interface {
	Run(context.Context) error
}

type app struct {
	c       config.Config
	s       storage.Storage
	spotify spotify.Spotify
	tg      telegram.Telegram
}

func New(c config.Config, s storage.Storage, spotify spotify.Spotify, tg telegram.Telegram) App {
	return &app{c, s, spotify, tg}
}

func (a app) Run(ctx context.Context) error {
	refreshToken, err := a.s.GetRefreshToken()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if refreshToken == "" {
		if err := a.spotify.Login(); err != nil {
			return err
		}
	} else {
		if err := a.spotify.RefreshAccessToken(refreshToken); err != nil {
			return err
		}
	}

	if err := a.tg.Login(); err != nil {
		return err
	}

	go GetTrackAndUpdateBio(ctx, a.spotify, a.tg)
	go UpdateAccessAndRefreshToken(ctx, a.spotify)

	<-ctx.Done()
	return nil
}

func UpdateAccessAndRefreshToken(ctx context.Context, s spotify.Spotify) {
	ticker := time.NewTicker(45 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.RefreshAccessToken(); err != nil {
				log.Printf("cant update refresh token %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func GetTrackAndUpdateBio(ctx context.Context, s spotify.Spotify, t telegram.Telegram) {
	var prev string
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			track, err := s.GetCurrentlyPlaying()
			if err != nil {
				log.Printf("Error getting track, waiting 10 seconds: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			if prev != track.String() {
				prev = track.String()
				err = t.UpdateBio(track.String())
				if err != nil {
					log.Printf("Error updating bio, waiting 10 seconds: %v", err)
					time.Sleep(10 * time.Second)
				}
			}
		case <-ctx.Done():
			return
		}

	}
}

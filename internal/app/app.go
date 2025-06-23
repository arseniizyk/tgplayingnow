package app

import (
	"log"
	"os"
	"time"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/spotify"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
	"github.com/arseniizyk/tgplayingnow/pkg/telegram"
)

type App interface {
	Run() error
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

func (a app) Run() error {
	refreshToken, err := a.s.GetRefreshToken()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if refreshToken == "" {
		err := a.spotify.Login()
		if err != nil {
			return err
		}
	} else {
		err := a.spotify.RefreshAccessToken(refreshToken)
		if err != nil {
			return err
		}
	}

	err = a.tg.Login()
	if err != nil {
		return err
	}

	go GetTrackAndUpdateBio(a.spotify, a.tg)
	go func() {
		for {
			time.Sleep(30 * time.Minute) // can be 60 min
			err := a.spotify.RefreshAccessToken()
			log.Printf("Cant update refresh token %v", err)
		}
	}()
	select {}
}

func GetTrackAndUpdateBio(s spotify.Spotify, t telegram.Telegram) {
	var prev string
	for {
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

		time.Sleep(15 * time.Second)
	}
}

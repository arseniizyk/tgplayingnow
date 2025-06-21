package app

import (
	"os"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/spotify"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
	"github.com/arseniizyk/tgplayingnow/pkg/telegram"
)

type App interface {
	Run() error
}

type app struct {
}

func New() App {
	return app{}
}

func (a app) Run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	storage, err := storage.New()
	if err != nil {
		return err
	}

	spotify := spotify.New(cfg, storage)
	if err := spotify.Login(); err != nil {
		return err
	}

	telegram := telegram.New(cfg)
	err = telegram.Login()
	if err != nil {
		return err
	}

	refreshToken, err := storage.GetRefreshToken()
	if err != nil {
		if os.IsNotExist(err) {
			err := spotify.Login()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		spotify.RefreshAccessToken()
	}

	return nil

	// for {
	// 	track, err := spotify.GetCurrentlyPlaying()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.Println(track.String())
	// 	time.Sleep(15 * time.Second)
	// }

}

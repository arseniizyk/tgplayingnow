package spotify

import (
	"errors"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/arseniizyk/tgplayingnow/pkg/storage"
)

type TrackResponse struct {
	Item struct {
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Name string `json:"name"`
	} `json:"item"`
}

type Spotify struct {
	storage      storage.Storage
	cfg          config.Config
	refreshToken string
	accessToken  string
}

var (
	ErrTokenExpired         = errors.New("token is bad or expired")
	ErrBadOauth             = errors.New("bad OAuth request")
	ErrRateLimits           = errors.New("too many requests, probably rate limits")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNoContent            = errors.New("not playing now")
)

func New(cfg config.Config, storage storage.Storage) *Spotify {
	return &Spotify{
		storage: storage,
		cfg:     cfg,
	}
}

package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arseniizyk/tgplayingnow/pkg/spotify/utils"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type Spotify interface {
	Login() error
	GetCurrentlyPlaying() (*strings.Builder, error)
	RefreshAccessToken() error
}

const (
	scope       = "user-read-currently-playing"
	redirectUri = "http://127.0.0.1:8080/spotify"
	authURL     = "https://accounts.spotify.com/authorize"
	tokenURL    = "https://accounts.spotify.com/api/token"
)

func (s *spotify) Login() error {
	verifier, challenge := utils.GenerateCodeVerifierAndChallenge()
	cfg := &oauth2.Config{
		ClientID:    s.cfg.SpotifyClientID(),
		RedirectURL: redirectUri,
		Scopes:      []string{scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	authURL := utils.BuildAuthURL(cfg, challenge)

	browser.OpenURL(authURL)

	token, err := handleCallback(cfg, verifier)
	if err != nil {
		return err
	}

	s.accessToken = token.AccessToken
	s.refreshToken = token.RefreshToken

	return nil
}

func (s *spotify) RefreshAccessToken() error {
	body := url.Values{}
	body.Add("grant_type", "refresh_token")
	body.Add("refresh_token", s.refreshToken)
	body.Add("client_id", s.cfg.SpotifyClientID())

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to refresh token %s", string(respBody))
	}

	var token oauth2.Token

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}

	if token.RefreshToken != "" {
		s.refreshToken = token.RefreshToken
	}

	s.accessToken = token.AccessToken

	return nil
}

func (s *spotify) GetCurrentlyPlaying() (*strings.Builder, error) {
	endpoint := "https://api.spotify.com/v1/me/player/currently-playing"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var tr TrackResponse
		err := json.NewDecoder(resp.Body).Decode(&tr)
		if err != nil {
			return nil, err
		}

		track := utils.FormatTrack(tr.Item.Name, tr.Item.Artists)
		return track, nil

	case http.StatusUnauthorized:
		return nil, ErrTokenExpired
	case http.StatusForbidden:
		return nil, ErrBadOauth
	case http.StatusTooManyRequests:
		return nil, ErrRateLimits
	default:
		log.Println(resp.StatusCode)
		return nil, ErrUnexpectedStatusCode
	}
}

func handleCallback(cfg *oauth2.Config, verifier string) (*oauth2.Token, error) {
	codeCh := make(chan string)

	mux := http.NewServeMux()
	mux.HandleFunc("/spotify", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "You can close this tab")
		codeCh <- code
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	code := <-codeCh

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		return nil, err
	}

	token, err := utils.TokenExchange(cfg, code, verifier)
	if err != nil {
		return nil, err
	}

	return token, nil
}

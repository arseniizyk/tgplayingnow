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

const (
	scope       = "user-read-currently-playing"
	redirectUri = "http://127.0.0.1:8080/spotify"
	authURL     = "https://accounts.spotify.com/authorize"
	tokenURL    = "https://accounts.spotify.com/api/token"
)

func (s *Spotify) Login() error {
	verifier, challenge := GenerateCodeVerifierAndChallenge()
	cfg := &oauth2.Config{
		ClientID:    s.cfg.SpotifyClientID(),
		RedirectURL: redirectUri,
		Scopes:      []string{scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	authURL := BuildAuthURL(cfg, challenge)

	err := browser.OpenURL(authURL)
	if err != nil {
		log.Printf("cant open browser %v, try to open it manually %s", err, authURL)
	}

	token, err := handleCallback(cfg, verifier)
	if err != nil {
		return err
	}

	s.accessToken = token.AccessToken
	s.refreshToken = token.RefreshToken
	if err := s.storage.SaveRefreshToken(s.refreshToken); err != nil {
		log.Println("cant save refresh token", err)
		return err
	}

	return nil
}

func (s *Spotify) RefreshAccessToken(refreshToken ...string) error {
	if refreshToken != nil && refreshToken[0] != "" {
		s.refreshToken = refreshToken[0]
	}

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
	defer utils.Dclose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to refresh token %s", string(respBody))
	}

	var token oauth2.Token

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}

	if token.RefreshToken != "" {
		log.Println("refresh token was updated")
		s.refreshToken = token.RefreshToken
		if err := s.storage.SaveRefreshToken(s.refreshToken); err != nil {
			log.Println("cant update refresh token", err)
			return err
		}
	}

	s.accessToken = token.AccessToken
	log.Println("access token was updated")

	return nil
}

func (s *Spotify) GetCurrentlyPlaying() (*strings.Builder, error) {
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
	defer utils.Dclose(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var tr TrackResponse
		err := json.NewDecoder(resp.Body).Decode(&tr)
		if err != nil {
			return nil, err
		}

		track := FormatTrack(tr.Item.Name, tr.Item.Artists)
		log.Println("Playing now:", track)
		return track, nil

	case http.StatusUnauthorized:
		return nil, ErrTokenExpired
	case http.StatusForbidden:
		return nil, ErrBadOauth
	case http.StatusTooManyRequests:
		return nil, ErrRateLimits
	case http.StatusNoContent:
		return nil, ErrNoContent
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
		if _, err := fmt.Fprintln(w, "You can close this tab"); err != nil {
			log.Println(err)
		}
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

	token, err := TokenExchange(cfg, code, verifier)
	if err != nil {
		return nil, err
	}

	return token, nil
}

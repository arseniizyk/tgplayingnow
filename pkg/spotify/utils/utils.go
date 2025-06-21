package utils

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

const maxBioLength = 70

func GenerateCodeVerifierAndChallenge() (string, string) {
	verifier := oauth2.GenerateVerifier()
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge
}

func BuildAuthURL(cfg *oauth2.Config, challenge string) string {
	return cfg.AuthCodeURL("state", oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)
}

func TokenExchange(cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
	ctx := context.Background()
	return cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
}

func FormatTrack(name string, artists []struct {
	Name string `json:"name"`
}) *strings.Builder {
	allArtists := make([]string, len(artists))

	for i, artist := range artists {
		allArtists[i] = artist.Name
	}

	builder := &strings.Builder{}

	for i := len(allArtists); i >= 1; i-- {
		joined := strings.Join(allArtists, ", ")
		result := fmt.Sprintf("%s - %s", name, joined)

		if len(result) <= maxBioLength {
			builder.WriteString(result)
			return builder
		}
	}

	builder.WriteString(name)
	return builder
}

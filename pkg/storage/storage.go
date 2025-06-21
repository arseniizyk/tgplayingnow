package storage

import (
	"os"
	"path/filepath"
)

type Storage interface {
	GetRefreshToken() (string, error)
	SaveRefreshToken(token string) error
}

type storage struct {
	filePath string
}

func New() (Storage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(configDir, "tgplayingnow")
	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return nil, err
	}

	tokenPath := filepath.Join(appDir, "refresh_token.txt")
	return &storage{filePath: tokenPath}, nil
}

func (s *storage) GetRefreshToken() (string, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *storage) SaveRefreshToken(token string) error {
	return os.WriteFile(s.filePath, []byte(token), 0644)
}

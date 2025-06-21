package telegram

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/arseniizyk/tgplayingnow/internal/config"
	"github.com/skip2/go-qrcode"
	"github.com/zelenin/go-tdlib/client"
)

type Telegram interface {
	Login() error
	UpdateBio(text string) error
}

type telegram struct {
	cfg    config.Config
	OldBio string
	client *client.Client
}

func New(cfg config.Config) Telegram {
	return &telegram{
		cfg: cfg,
	}
}

func (t *telegram) Login() error {
	params := &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     false,
		UseChatInfoDatabase: false,
		UseMessageDatabase:  false,
		UseSecretChats:      false,
		ApiId:               t.cfg.TelegramAppId(),
		ApiHash:             t.cfg.TelegramAppHash(),
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}

	authorizer := client.QrAuthorizer(params, func(link string) error {
		return qrcode.WriteFile(link, qrcode.Medium, 256, "qr.png")
	})

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		return fmt.Errorf("NewClient error %s", err)
	}

	t.client = tdlibClient
	u, _ := t.client.GetMe()

	info, err := t.client.GetUserFullInfo(&client.GetUserFullInfoRequest{UserId: u.Id})
	if err != nil {
		log.Println("cant get previous bio", err)
	}

	t.OldBio = info.Bio.Text

	return nil
}

func (t *telegram) UpdateBio(text string) error {
	_, err := t.client.SetBio(&client.SetBioRequest{Bio: text})
	return err
}

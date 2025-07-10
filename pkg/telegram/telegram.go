// nolint
package telegram

import (
	"fmt"
	"log"
	"os"

	"github.com/skip2/go-qrcode"
	"github.com/zelenin/go-tdlib/client" //nolint:all
)

func (t *Telegram) Login() error {
	params := generateParams(t.cfg.TelegramAppId(), t.cfg.TelegramAppHash())

	authorizer := client.QrAuthorizer(params, func(link string) error {
		err := qrcode.WriteFile(link, qrcode.Medium, 256, "qr.png")
		if err != nil {
			return fmt.Errorf("failed to write QR code: %w", err)
		}

		if err := openFile("./qr.png"); err != nil {
			log.Printf("failed to open QR code image: %v", err)
		}

		return nil
	})

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		return fmt.Errorf("NewClient error %s", err)
	}

	defer os.Remove("./qr.png")

	t.c = tdlibClient
	u, err := t.c.GetMe()
	if err != nil {
		return fmt.Errorf("GetMe error: %w", err)
	}

	log.Printf("%s | %s | [%s] \n", u.FirstName, u.LastName, u.Usernames.ActiveUsernames)

	info, err := t.c.GetUserFullInfo(&client.GetUserFullInfoRequest{UserId: u.Id})
	if err != nil {
		log.Println("cant get previous bio", err)
	} else {
		log.Println(info.Bio.Text)
		t.oldBio = info.Bio.Text
	}

	return nil
}

func (t *Telegram) UpdateBio(text string) error {
	_, err := t.c.SetBio(&client.SetBioRequest{Bio: text})
	log.Println("Updating bio:", text)
	return err
}

func (t *Telegram) ResetBio() error {
	_, err := t.c.SetBio(&client.SetBioRequest{Bio: t.oldBio})
	if err != nil {
		return fmt.Errorf("Error returning old bio")
	}

	log.Printf("Old bio %s was returned\n", t.oldBio)
	return nil
}

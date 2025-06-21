package main

import (
	"github.com/arseniizyk/tgplayingnow/internal/app"
)

func main() {
	app := app.New()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

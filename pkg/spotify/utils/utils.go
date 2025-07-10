package utils

import (
	"io"
	"log"
)

func Dclose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println("Cant close response body", err)
	}
}

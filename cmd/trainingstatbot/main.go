package main

import (
	"github.com/vlladoff/trainingstatbot/internal/app/trainingstatbot"
	"log"
)

func main() {
	if err := trainingstatbot.Start(); err != nil {
		log.Fatal(err)
	}
}

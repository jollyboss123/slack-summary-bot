package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	botToken := os.Getenv("SLACK_TOKEN")
	if botToken == "" {
		log.Fatal("SLACK_TOKEN is required")
		return
	}

	client := slack.New(botToken)

	channelId := "#test"
	_, _, err = client.PostMessage(channelId, slack.MsgOptionText("Hello, world!", false))
	if err != nil {
		log.Fatal(err)
		return
	}
}

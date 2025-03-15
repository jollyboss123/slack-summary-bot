package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	logger := log.New(os.Stdout, "slack-bot: ", log.LstdFlags|log.Lshortfile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		logger.Fatal("SLACK_APP_TOKEN is required")
		os.Exit(1)
		return
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		logger.Fatal("SLACK_APP_TOKEN should start with xapp-")
		os.Exit(1)
		return
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("SLACK_TOKEN is required")
		os.Exit(1)
		return
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		logger.Fatal("SLACK_BOT_TOKEN should start with xoxb-")
		os.Exit(1)
		return
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(appToken),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.LstdFlags|log.Lshortfile)),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.LstdFlags|log.Lshortfile)),
	)

	socketmodeHandler := socketmode.NewSocketmodeHandler(client)

	socketmodeHandler.HandleSlashCommand("/summarize", func(evt *socketmode.Event, clt *socketmode.Client) {
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			logger.Printf("Error converting event to Slash command: %+v", evt)
			return
		}

		clt.Ack(*evt.Request)

		_, _, err := client.PostMessage(cmd.ChannelID, slack.MsgOptionText("Hello, world!", false))
		if err != nil {
			logger.Printf("Error posting message: %v", err)
			return
		}
	})

	socketmodeHandler.RunEventLoop()
}

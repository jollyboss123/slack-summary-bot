package main

import (
	"log"
	"os"
	"strconv"
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

		if cmd.Text != "" && !strings.HasPrefix(cmd.Text, "--") {
			clt.Ack(*evt.Request, map[string]any{
				"text": "Please provide a valid flag, use /summarize --help for more information",
			})
			return
		}

		if cmd.Text == "" || cmd.Text == "--to-channel" {
			clt.Ack(*evt.Request)

			msg, err := getChannelHistory(api, cmd.ChannelID)
			if err != nil {
				logger.Printf("Error getting channel history: %v", err)
				return
			}

			_, _, err = client.PostMessage(cmd.ChannelID, slack.MsgOptionText(strconv.Itoa(len(msg)), false))
			if err != nil {
				logger.Printf("Error posting message: %v", err)
				return
			}
			return
		}

		if cmd.Text == "--to-me" {
			msg, err := getChannelHistory(api, cmd.ChannelID)
			if err != nil {
				logger.Printf("Error getting channel history: %v", err)
				return
			}

			clt.Ack(*evt.Request, map[string]any{
				"text": strconv.Itoa(len(msg)),
			})
			return
		}

		if cmd.Text == "--help" {
			clt.Ack(*evt.Request, map[string]any{
				"text": `Usage: /summarize [flag]
--to-channel: Send message to channel
--to-me: Send message to me
`,
			})
			return
		}

		clt.Ack(*evt.Request, map[string]any{
			"text": "Invalid flag, use /summarize --help for more information",
		})
	})

	socketmodeHandler.RunEventLoop()
}

func getChannelHistory(api *slack.Client, channelID string) ([]slack.Message, error) {
	history, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, err
	}

	return history.Messages, nil

}

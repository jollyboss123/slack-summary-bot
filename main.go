package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// TODO: add caching for summary with channel id as key
// TODO: restrict tokens use per call to openai to avoid abuse
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

	slackClient := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.LstdFlags|log.Lshortfile)),
	)

	openaiClient := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	socketmodeHandler := socketmode.NewSocketmodeHandler(slackClient)

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

			msg, err := getChannelSummary(api, openaiClient, cmd.ChannelID)
			if err != nil {
				logger.Printf("Error getting channel summary: %v", err)
				return
			}

			_, _, err = clt.PostMessage(cmd.ChannelID, slack.MsgOptionText(msg, false))
			if err != nil {
				logger.Printf("Error posting message: %v", err)
				return
			}
			return
		}

		if cmd.Text == "--to-me" {
			msg, err := getChannelSummary(api, openaiClient, cmd.ChannelID)
			if err != nil {
				logger.Printf("Error getting channel summary: %v", err)
				return
			}

			clt.Ack(*evt.Request, map[string]any{
				"text": msg,
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

func getChannelSummary(sc *slack.Client, oac *openai.Client, channelID string) (string, error) {
	history, err := getChannelHistory(sc, channelID)
	if err != nil {
		return "", err
	}

	return getSummary(oac, history, "Summarize the following Slack conversation:", "Make sure to keep it concise and professional.")
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

func getSummary(clt *openai.Client, messages []slack.Message, prePrompt, postPrompt string) (string, error) {
	var chatMessages []openai.ChatCompletionMessageParamUnion

	if prePrompt != "" {
		chatMessages = append(chatMessages, openai.SystemMessage(
			prePrompt,
		))
	}

	for _, msg := range messages {
		chatMessages = append(chatMessages, openai.UserMessage(
			msg.Text,
		))
	}

	if postPrompt != "" {
		chatMessages = append(chatMessages, openai.SystemMessage(
			postPrompt,
		))
	}

	cc, err := clt.Chat.Completions.New(
		context.TODO(),
		openai.ChatCompletionNewParams{
			Model:    openai.F(openai.ChatModelGPT4oMini),
			Messages: openai.F(chatMessages),
		},
	)
	if err != nil {
		return "", err
	}

	return cc.Choices[0].Message.Content, nil
}

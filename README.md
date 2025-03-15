# Slack Summary Bot

## Overview
Slack Summary Bot is a Golang-based Slack bot that summarizes messages in a Slack channel using OpenAI's GPT models. It utilizes Slack's Socket Mode for real-time interactions and the `slack-go/slack` SDK for API communication.

## Features
- Retrieves channel message history
- Generates a summary using OpenAI's GPT models
- Supports custom pre-prompts and post-prompts
- Uses Slack Socket Mode for seamless real-time interaction

## Installation

### Prerequisites
Ensure you have the following installed:
- **Go (>=1.18)**
- A Slack workspace with bot permissions
- An OpenAI API key
- A `.env` file with required credentials

### Clone the Repository
```bash
git clone https://github.com/jollyboss123/slack-summary-bot.git
cd slack-summary-bot
```

### Install Dependencies
```bash
go mod tidy
```

## Configuration
Create a `.env` file in the root directory with:
```ini
SLACK_APP_TOKEN=xapp-...
SLACK_BOT_TOKEN=xoxb-...
OPENAI_API_KEY=sk-...
```

## Running the Bot
```bash
go run main.go
```

## Usage
- Mention the bot in a Slack channel and use a slash command (e.g., `/summarize`)
- The bot fetches recent messages and generates a summary
- Supports optional custom prompts before and after the summary

## Deployment
You can deploy the bot using Docker:
```bash
docker build -t slack-summary-bot .
docker run --env-file .env slack-summary-bot
```

## Troubleshooting
### 1. **Quota Issues with OpenAI**
If you get a `429 Too Many Requests` error, check your [OpenAI billing](https://platform.openai.com/account/billing).

### 2. **Bot Not Responding in Slack**
- Ensure the bot is **installed in the workspace**
- Verify **Socket Mode is enabled** in Slack app settings

## License
This project is licensed under the MIT License.

## Contributing
Pull requests are welcome! Open an issue for discussions.

---
Happy coding! 🚀


package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	BotToken    string `envconfig:"BOT_TOKEN" required:"true"`
	WebhookURL  string `envconfig:"WEBHOOK_URL" required:"true"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"443"`
}

type HerokuConfig struct {
	Port string `envconfig:"PORT" default:"5000"`
}

func main() {
	var config Configuration
	err := envconfig.Process("CLB", &config)
	if err != nil {
		logrus.Fatalf("Failed to process config: %s", err)
	}
	var herokuConfig HerokuConfig
	err = envconfig.Process("", &herokuConfig)
	if err != nil {
		logrus.Fatalf("Failed to parse Heroku config: %s", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		logrus.Fatalf("Failed to create bot: %s", err)
	}
	logrus.Infof("Authorised! Bot username: %s", bot.Self.UserName)

	webhook, err := url.Parse(config.WebhookURL)
	if err != nil {
		logrus.Fatalf("Could not parse webhook URL: %s", err)
	}
	// set port in webhook
	webhook.Host = fmt.Sprintf("%s:%s", webhook.Hostname(), config.WebhookPort)

	// add the bot token to the URL
	webhook.Path = filepath.Join(webhook.Path, config.BotToken)

	logrus.Infof("Attempting to set webhook to '%s'", webhook.String())
	// TODO: potentially use NewWebhookWithCert
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(webhook.String()))
	if err != nil {
		logrus.Fatalf("Failed to setup webhook: %s", err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		logrus.Fatalf("Failed to get webhook info: %s", err)
	}
	if info.LastErrorDate != 0 {
		logrus.Fatalf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook(webhook.Path)
	logrus.Infof("Set webook; will listen on '%s'", webhook.Path)
	// TODO: potentially use ListenAndServeTLS
	serveOn := fmt.Sprintf("%s:%s", "0.0.0.0", herokuConfig.Port)
	go func() {
		logrus.Infof("Serving webhook on %s", serveOn)
		if err := http.ListenAndServe(serveOn, nil); err != nil {
			logrus.Fatalf("http listener exited with error: %s", err)
		}
		logrus.Fatal("http listener exited without error")
	}()

	for update := range updates {
		// do something
		logrus.WithFields(logrus.Fields{
			"update": update,
		}).Info("Received update")
		if update.Message == nil {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"ID":       update.Message.MessageID,
			"chat_ID":  update.Message.Chat.ID,
			"user_ID":  update.Message.From.ID,
			"username": update.Message.From.UserName,
			"content":  update.Message.Text,
		}).Info("Received message")
	}
}

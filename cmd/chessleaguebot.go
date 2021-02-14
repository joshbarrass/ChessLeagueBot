package main

import (
	"fmt"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	BotToken   string `envconfig:"BOT_TOKEN" required:"true"`
	WebhookURL string `envconfig:"WEBHOOK_URL" required:"true"`
	ListenAddr string `envconfig:"LISTEN_ADDRESS" default:"0.0.0.0"`
	ListenPort string `envconfig:"LISTEN_PORT" default:"8080"`
}

func main() {
	var config Configuration
	err := envconfig.Process("CLB", &config)
	if err != nil {
		logrus.Fatalf("Failed to process config: %s", err)
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
	// check if port exists in webhook and update config if so
	if webhookPort := webhook.Port(); webhookPort != "" {
		if config.ListenPort != webhookPort && config.ListenPort != "8080" {
			logrus.Fatalf("Differing ports were specified: %s and %s", webhookPort, config.ListenPort)
		}
		config.ListenPort = webhookPort
	}
	// set the port in the listen address
	webhook.Host = fmt.Sprintf("%s:%s", webhook.Hostname(), config.ListenPort)

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

	updates := bot.ListenForWebhook(webhook.String())
	logrus.Infof("Set webook; will listen on URL '%s'", webhook.String())
	// TODO: potentially use ListenAndServeTLS
	go http.ListenAndServe(fmt.Sprintf("%s:%s", config.ListenAddr, config.ListenPort), nil)
	logrus.Info("Serving on webhook...")

	for update := range updates {
		// do something
		logrus.WithField(
			"update", update,
		).Info("Received update")
	}
}

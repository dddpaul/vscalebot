package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vscale/go-vscale"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"time"
)

var (
	botCmd = &cobra.Command{
		Use:   "bot",
		Short: "Telegram Bot",
		Run: func(cmd *cobra.Command, args []string) {
			if len(telegramToken) == 0 {
				log.Panic("Telegram API token has to be specified")
			}

			if len(accountsStrings) == 0 {
				log.Panic("At least one Vscale account map has to be specified")
			}

			accountsFlags := arrayFlags(accountsStrings)
			accounts, err := accountsFlags.toMap()
			if err != nil {
				log.Panic(err)
			}

			startBot(accounts)
		},
	}
)

func init() {
	rootCmd.AddCommand(botCmd)
}

func startBot(accounts []*VscaleAccount) {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = verbose
	log.Printf("Authorized on account %s", bot.Self.UserName)

	c := make(chan tgbotapi.MessageConfig)
	go func() {
		for msg := range c {
			bot.Send(msg)
		}
	}()

	subscribed := false
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {
			if subscribed {
				for _, acc := range accounts {
					c <- tgbotapi.NewMessage(acc.ID, fmt.Sprintf("%s balance is %.2f roubles", acc.Name, balance(acc.Token)))
				}
			}
		}
	}()

	updates, _ := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   0,
		Timeout: 60,
	})

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		log.Printf("[%s] %s", update.Message.From.UserName, text)

		for _, acc := range accounts {
			acc.ID = update.Message.Chat.ID
			if text == "/"+acc.Name {
				c <- tgbotapi.NewMessage(acc.ID, fmt.Sprintf("%s balance is %.2f roubles", acc.Name, balance(acc.Token)))
			} else if text == "/start" {
				subscribed = true
				c <- tgbotapi.NewMessage(acc.ID, fmt.Sprintf("%s subscribed", acc.Name))
			} else if text == "/stop" {
				subscribed = false
				c <- tgbotapi.NewMessage(acc.ID, fmt.Sprintf("%s unsubscribed", acc.Name))
			}
		}
	}
}

func balance(token string) float64 {
	client := vscale_api_go.NewClient(token)
	billing, _, err := client.Billing.Billing()
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
	return float64(billing.Balance) / 100
}

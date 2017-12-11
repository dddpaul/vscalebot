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

		for _, account := range accounts {
			if text == "/"+account.Name {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("%s balance is %.2f roubles", account.Name, balance(account.Token)))
				bot.Send(msg)
			} else if text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Subscription accepted"))
				bot.Send(msg)
				go func() {
					ticker := time.NewTicker(time.Minute * 1)
					for range ticker.C {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("%s balance is %.2f roubles", account.Name, balance(account.Token)))
						bot.Send(msg)
					}
				}()
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

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/vscale/go-vscale"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"strings"
	"time"
)

type VscaleAccount struct {
	Name  string
	Token string
	ID    int64
}

type arrayFlags []string

func (flags *arrayFlags) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) toMap() ([]*VscaleAccount, error) {
	var accounts []*VscaleAccount
	for _, s := range *flags {
		items := strings.Split(s, "=")
		if len(items) == 0 {
			return nil, errors.New("incorrect Vscale name to token map format")
		}
		accounts = append(accounts, &VscaleAccount{
			Name:  items[0],
			Token: items[1],
		})
	}
	return accounts, nil
}

var verbose bool
var telegramToken string
var accountsFlags arrayFlags

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.StringVar(&telegramToken, "telegram-token", "", "Telegram API token")
	flag.Var(&accountsFlags, "vscale", "List of Vscale name to token maps, i.e. 'swarm=123456'")
	flag.Parse()

	if len(telegramToken) == 0 {
		log.Panic("Telegram API token has to be specified")
	}

	if len(accountsFlags) == 0 {
		log.Panic("At least one Vscale account map has to be specified")
	}

	accounts, err := accountsFlags.toMap()
	if err != nil {
		log.Panic(err)
	}

	start(accounts)
}

func start(accounts []*VscaleAccount) {
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

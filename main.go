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
	Token  string
	ChatID int64
}

type arrayFlags []string

func (flags *arrayFlags) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) toMap() (map[string]*VscaleAccount, error) {
	accounts := make(map[string]*VscaleAccount)
	for _, s := range *flags {
		items := strings.Split(s, "=")
		if len(items) == 0 {
			return nil, errors.New("incorrect Vscale name to token map format")
		}
		accounts[items[0]] = &VscaleAccount{
			Token: items[1],
		}
	}
	return accounts, nil
}

var (
	verbose       bool
	telegramToken string
	accountsFlags arrayFlags
	interval      time.Duration
	threshold     float64
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.StringVar(&telegramToken, "telegram-token", "", "Telegram API token")
	flag.Var(&accountsFlags, "vscale", "List of Vscale name to token maps, i.e. 'swarm=123456'")
	flag.DurationVar(&interval, "interval", 600000000000, "Subscription messages interval in nanoseconds")
	flag.Float64Var(&threshold, "threshold", 100, "Subscription messages threshold in roubles")
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

func start(accounts map[string]*VscaleAccount) {
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
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if subscribed {
				for name, acc := range accounts {
					balance := balance(acc.Token)
					if balance <= threshold {
						c <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s balance is %.2f roubles", name, balance))
					}
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

		for name, acc := range accounts {
			acc.ChatID = update.Message.Chat.ID
			if text == "/"+name {
				c <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s balance is %.2f roubles", name, balance(acc.Token)))
			} else if text == "/start" {
				subscribed = true
				c <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, threshold))
			} else if text == "/stop" {
				subscribed = false
				c <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s unsubscribed", name))
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

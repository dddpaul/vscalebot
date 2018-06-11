package telegram

import (
	"fmt"
	"log"
	"time"

	"github.com/dddpaul/vscalebot/vscale"
	"github.com/docker/libkv/store"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	bot       *tb.Bot
	accounts  map[string]*vscale.Account
	threshold float64
	interval  time.Duration
	kvStore   store.Store
	verbose   bool
}

func NewBot(telegramToken string, accounts map[string]*vscale.Account, threshold float64, interval time.Duration, kvStore store.Store, verbose bool) (*Bot, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Me.Username)

	subscribed := false
	for name, acc := range accounts {
		bot.Handle("/"+name, func(m *tb.Message) {
			bot.Send(m.Sender, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
		})
		bot.Handle("/start", func(m *tb.Message) {
			subscribed = true
			bot.Send(m.Sender, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, threshold))
		})
		bot.Handle("/stop", func(m *tb.Message) {
			subscribed = false
			bot.Send(m.Sender, fmt.Sprintf("%s unsubscribed", name))
		})
	}

	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if subscribed {
				for name, acc := range accounts {
					balance := vscale.Balance(acc.Token)
					if balance <= threshold {
						bot.Send(acc, fmt.Sprintf("%s balance is %.2f roubles", name, balance))
					}
				}
			}
		}
	}()

	return &Bot{
		bot:       bot,
		accounts:  accounts,
		threshold: threshold,
		interval:  interval,
		kvStore:   kvStore,
		verbose:   verbose,
	}, nil
}

func (bot *Bot) Start() {
	bot.bot.Start()
}

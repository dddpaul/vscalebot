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

type BotOption func(b *Bot)

func WithThreshold(t float64) BotOption {
	return func(b *Bot) {
		b.threshold = t
	}
}

func WithInterval(i time.Duration) BotOption {
	return func(b *Bot) {
		b.interval = i
	}
}

// func NewBot(telegramToken string, accounts map[string]*vscale.Account, threshold float64, interval time.Duration, kvStore store.Store, verbose bool) (*Bot, error) {
func NewBot(telegramToken string, accounts map[string]*vscale.Account, opts ...BotOption) (*Bot, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Me.Username)

	subscribed := false
	bot.Handle("/balance", func(m *tb.Message) {
		for name, acc := range accounts {
			bot.Send(m.Sender, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
		}
	})
	bot.Handle("/start", func(m *tb.Message) {
		subscribed = true
		for name := range accounts {
			bot.Send(m.Sender, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, threshold))
		}
	})
	bot.Handle("/stop", func(m *tb.Message) {
		subscribed = false
		for name := range accounts {
			bot.Send(m.Sender, fmt.Sprintf("%s unsubscribed", name))
		}
	})

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

	for _, opt := range opts {
		opt(bot)
	}
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

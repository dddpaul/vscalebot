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
	store     store.Store
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

func WithStore(s store.Store) BotOption {
	return func(b *Bot) {
		b.store = s
	}
}

func WithVerbose(v bool) BotOption {
	return func(b *Bot) {
		b.verbose = v
	}
}

func NewBot(telegramToken string, accounts map[string]*vscale.Account, opts ...BotOption) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s\n", b.Me.Username)

	bot := &Bot{
		bot:       b,
		accounts:  accounts,
		threshold: 0,
		interval:  0,
		store:     nil,
		verbose:   false,
	}

	for _, opt := range opts {
		opt(bot)
	}

	subscribed := false
	b.Handle("/balance", func(m *tb.Message) {
		for name, acc := range accounts {
			b.Send(m.Sender, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
		}
	})
	b.Handle("/start", func(m *tb.Message) {
		subscribed = true
		for name := range accounts {
			b.Send(m.Sender, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, bot.threshold))
		}
	})
	b.Handle("/stop", func(m *tb.Message) {
		subscribed = false
		for name := range accounts {
			b.Send(m.Sender, fmt.Sprintf("%s unsubscribed", name))
		}
	})

	go func() {
		ticker := time.NewTicker(bot.interval)
		for range ticker.C {
			if subscribed {
				for name, acc := range accounts {
					balance := vscale.Balance(acc.Token)
					if balance <= bot.threshold {
						b.Send(acc, fmt.Sprintf("%s balance is %.2f roubles", name, balance))
					}
				}
			}
		}
	}()

	return bot, nil
}

func (b *Bot) Start() {
	b.bot.Start()
}

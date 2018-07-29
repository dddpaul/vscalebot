package telegram

import (
	"fmt"
	"log"
	"time"

	"github.com/dddpaul/vscalebot/pkg/vscale"
	"github.com/docker/libkv/store"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BotChatStore interface {
	List() ([]tb.Chat, error)
	Add(tb.Chat) error
	Remove(tb.Chat) error
}

type Bot struct {
	bot       *tb.Bot
	chats     BotChatStore
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

func WithVerbose(v bool) BotOption {
	return func(b *Bot) {
		b.verbose = v
	}
}

func NewBot(telegramToken string, chats BotChatStore, accounts map[string]*vscale.Account, opts ...BotOption) (*Bot, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s\n", bot.Me.Username)

	b := &Bot{
		bot:       bot,
		chats:     chats,
		accounts:  accounts,
		threshold: 0,
		interval:  0,
		verbose:   false,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

func (b *Bot) Start() {
	go func() {
		ticker := time.NewTicker(b.interval)
		for range ticker.C {
			chats, err := b.chats.List()
			if err != nil {
				log.Panic(err)
			}
			for _, c := range chats {
				for name, acc := range b.accounts {
					balance := vscale.Balance(acc.Token)
					if balance <= b.threshold {
						b.bot.Send(&c, fmt.Sprintf("%s balance is %.2f roubles", name, balance))
					}
				}
			}
		}
	}()

	b.bot.Handle("/balance", func(m *tb.Message) {
		for name, acc := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
		}
	})
	b.bot.Handle("/start", func(m *tb.Message) {
		b.chats.Add(*m.Chat)
		for name := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, b.threshold))
		}
	})
	b.bot.Handle("/stop", func(m *tb.Message) {
		b.chats.Remove(*m.Chat)
		for name := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s unsubscribed", name))
		}
	})

	b.bot.Start()
}

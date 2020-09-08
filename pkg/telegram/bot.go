package telegram

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dddpaul/vscalebot/pkg/vscale"
	"golang.org/x/net/proxy"
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
	verbose   bool
	admin     string
	client    *http.Client
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

func WithSocks(s string) BotOption {
	return func(b *Bot) {
		if len(s) == 0 {
			return
		}

		u, err := url.Parse(s)
		if err != nil {
			log.Panic(err)
		}

		var auth *proxy.Auth
		if u.User != nil {
			auth = &proxy.Auth{
				User: u.User.Username(),
			}
			if p, ok := u.User.Password(); ok {
				auth.Password = p
			}
		}

		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		if err != nil {
			log.Panic(err)
		}
		httpTransport := &http.Transport{
			Dial: dialer.Dial,
		}
		client := &http.Client{Transport: httpTransport}
		b.client = client
	}
}

func WithAdmin(a string) BotOption {
	return func(b *Bot) {
		b.admin = a
	}
}

func NewBot(telegramToken string, chats BotChatStore, accounts map[string]*vscale.Account, opts ...BotOption) (*Bot, error) {
	b := &Bot{
		chats:    chats,
		accounts: accounts,
	}

	for _, opt := range opts {
		opt(b)
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Client: b.client,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s\n", bot.Me.Username)

	b.bot = bot
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

	check := func(m *tb.Message) bool {
		if b.admin != "" && b.admin != m.Sender.Username {
			b.bot.Send(m.Sender, "Access restricted")
			return false
		}
		return true
	}

	b.bot.Handle("/balance", func(m *tb.Message) {
		if !check(m) {
			return
		}
		for name, acc := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
		}
	})
	b.bot.Handle("/start", func(m *tb.Message) {
		if !check(m) {
			return
		}
		b.chats.Add(*m.Chat)
		for name := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, b.threshold))
		}
	})
	b.bot.Handle("/stop", func(m *tb.Message) {
		if !check(m) {
			return
		}
		b.chats.Remove(*m.Chat)
		for name := range b.accounts {
			b.bot.Send(m.Sender, fmt.Sprintf("%s unsubscribed", name))
		}
	})
	b.bot.Handle("/status", func(m *tb.Message) {
		if !check(m) {
			return
		}
		chats, err := b.chats.List()
		if err != nil {
			log.Panic(err)
		}
		b.bot.Send(m.Sender, fmt.Sprintf("Subscribers: %d", len(chats)))
	})

	b.bot.Start()
}

package telegram

import (
	"fmt"
	"log"
	"time"

	"github.com/dddpaul/vscalebot/vscale"
	"github.com/docker/libkv/store"
	"gopkg.in/telegram-bot-api.v4"
)

type Bot struct {
	bot       *tgbotapi.BotAPI
	accounts  map[string]*vscale.Account
	threshold float64
	sendChan  chan tgbotapi.MessageConfig
	interval  time.Duration
	kvStore   store.Store
	verbose   bool
}

func NewBot(telegramToken string, accounts map[string]*vscale.Account, threshold float64, interval time.Duration, kvStore store.Store, verbose bool) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, err
	}

	bot.Debug = verbose
	log.Printf("Authorized on account %s", bot.Self.UserName)

	c := make(chan tgbotapi.MessageConfig)
	go func() {
		for msg := range c {
			bot.Send(msg)
		}
	}()

	return &Bot{
		bot:       bot,
		accounts:  accounts,
		threshold: threshold,
		sendChan:  c,
		interval:  interval,
		kvStore:   kvStore,
		verbose:   verbose,
	}, nil
}

func (bot *Bot) Run() {
	subscribed := false
	go func() {
		ticker := time.NewTicker(bot.interval)
		for range ticker.C {
			if subscribed {
				for name, acc := range bot.accounts {
					balance := vscale.Balance(acc.Token)
					if balance <= bot.threshold {
						bot.sendChan <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s balance is %.2f roubles", name, balance))
					}
				}
			}
		}
	}()

	updates, _ := bot.bot.GetUpdatesChan(tgbotapi.UpdateConfig{
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

		for name, acc := range bot.accounts {
			acc.ChatID = update.Message.Chat.ID
			if text == "/"+name {
				bot.sendChan <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s balance is %.2f roubles", name, vscale.Balance(acc.Token)))
			} else if text == "/start" {
				subscribed = true
				bot.sendChan <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s subscribed with %.2f roubles threshold", name, bot.threshold))
			} else if text == "/stop" {
				subscribed = false
				bot.sendChan <- tgbotapi.NewMessage(acc.ChatID, fmt.Sprintf("%s unsubscribed", name))
			}
		}
	}
}

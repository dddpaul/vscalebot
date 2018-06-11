package main

import (
	"errors"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/dddpaul/vscalebot/telegram"
	"github.com/dddpaul/vscalebot/vscale"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
)

type arrayFlags []string

func (flags *arrayFlags) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) toMap() (map[string]*vscale.Account, error) {
	accounts := make(map[string]*vscale.Account)
	for _, s := range *flags {
		items := strings.Split(s, "=")
		if len(items) == 0 {
			return nil, errors.New("incorrect Vscale name to token map format")
		}
		accounts[items[0]] = &vscale.Account{
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
	boltPath      string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.StringVar(&telegramToken, "telegram-token", "", "Telegram API token")
	flag.Var(&accountsFlags, "vscale", "List of Vscale name to token maps, i.e. 'swarm=123456'")
	flag.DurationVar(&interval, "interval", 600000000000, "Subscription messages interval in nanoseconds")
	flag.Float64Var(&threshold, "threshold", 100, "Subscription messages threshold in roubles")
	flag.StringVar(&boltPath, "bolt-path", "", "BoltDB path")
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

	kvStore, err := boltdb.New([]string{boltPath}, &store.Config{Bucket: "alertmanager"})
	if err != nil {
		log.Panic(err)
	}
	defer kvStore.Close()

	bot, err := telegram.NewBot(telegramToken, accounts, threshold, interval, kvStore, verbose)
	if err != nil {
		log.Panic(err)
	}

	bot.Run()
}

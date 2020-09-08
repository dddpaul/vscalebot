package main

import (
	"errors"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/dddpaul/vscalebot/pkg/telegram"
	"github.com/dddpaul/vscalebot/pkg/vscale"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
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
	verbose          bool
	telegramToken    string
	telegramProxyURL string
	telegramAdmin    string
	accountsFlags    arrayFlags
	interval         time.Duration
	threshold        float64
	storeType        string
	boltPath         string
	consulURL        string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.StringVar(&telegramToken, "telegram-token", "", "Telegram API token")
	flag.StringVar(&telegramProxyURL, "telegram-proxy-url", "", "Telegram SOCKS5 proxy url")
	flag.StringVar(&telegramAdmin, "telegram-admin", "", "Telegram admin user")
	flag.Var(&accountsFlags, "vscale", "List of Vscale name to token maps, i.e. 'swarm=123456'")
	flag.DurationVar(&interval, "interval", 600000000000, "Subscription messages interval in nanoseconds")
	flag.Float64Var(&threshold, "threshold", 100, "Subscription messages threshold in roubles")
	flag.StringVar(&storeType, "store", "bolt", "Store type - bolt / consul")
	flag.StringVar(&boltPath, "bolt-path", "", "BoltDB path")
	flag.StringVar(&consulURL, "consul-url", "", "Consul URL")
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

	var kv store.Store
	switch storeType {
	case "bolt":
		kv, err = boltdb.New([]string{boltPath}, &store.Config{Bucket: "alertmanager"})
		if err != nil {
			log.Panic(err)
		}
	case "consul":
		kv, err = consul.New([]string{consulURL}, nil)
		if err != nil {
			log.Panic(err)
		}
	default:
		log.Panicf("Store must be bolt or consul!\n")
	}
	defer kv.Close()

	chats, err := telegram.NewChatStore(kv)
	if err != nil {
		log.Panic(err)
	}

	bot, err := telegram.NewBot(telegramToken, chats, accounts,
		telegram.WithThreshold(threshold),
		telegram.WithInterval(interval),
		telegram.WithVerbose(verbose),
		telegram.WithAdmin(telegramAdmin),
		telegram.WithSocks(telegramProxyURL))
	if err != nil {
		log.Panic(err)
	}

	bot.Start()
}

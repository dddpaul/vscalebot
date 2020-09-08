Vscale Bot
=========

Simple Telegram bot for vscale.io written in Go.

Install:

```bash
go get -u github.com/dddpaul/vscalebot
```

Or grab Docker image:

```bash
docker pull dddpaul/vscalebot
```

Usage:

```bash
Usage of vscalebot:
  -bolt-path string
    	BoltDB path
  -consul-url string
    	Consul URL
  -interval duration
    	Subscription messages interval in nanoseconds (default 10m0s)
  -store string
    	Store type - bolt / consul (default "bolt")
  -telegram-proxy-url string
    	Telegram SOCKS5 proxy url
  -telegram-token string
    	Telegram API token
  -telegram-admin string
    	Telegram admin user
  -threshold float
    	Subscription messages threshold in roubles (default 100)
  -verbose
    	Enable bot debug
  -vscale value
    	List of Vscale name to token maps, i.e. 'swarm=123456'
```

Then for `/swarm` command Vscale balance will be shown in Telegram bot.

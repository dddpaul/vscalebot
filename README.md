Vscale Bot
=========

Simple Telegram bot for vscale.ru written in Go.

Install:

```
go get -u github.com/dddpaul/vscalebot
```

Or grab Docker image:

```
docker pull dddpaul/vscalebot
```

Usage:

```
Usage of ./vscalebot:
  -interval duration
    	Subscription messages interval in nanoseconds (default 10m0s)
  -telegram-token string
    	Telegram API token
  -threshold float
    	Subscription messages threshold in roubles (default 100)
  -verbose
    	Enable bot debug
  -vscale value
    	List of Vscale name to token maps, i.e. 'swarm=123456'
```

Then for `/swarm` command Vscale balance will be shown in Telegram bot.

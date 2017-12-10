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
  -telegram-token string
    	Telegram API token
  -verbose
    	Enable bot debug
  -vscale value
    	List of Vscale name to token maps, i.e. 'swarm=123456'
```

Then for `/swarm` command Vscale balance will be shown in Telegram bot.

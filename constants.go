package main

import (
	"os"

	"github.com/BiRabittoh/disgord/mylog"
	"github.com/kkdai/youtube/v2"
)

var (
	discordToken string
	appID        string
	prefix       string = "$"

	logger = mylog.NewLogger(os.Stdout, "main", mylog.DEBUG)
	yt     = youtube.Client{}
)

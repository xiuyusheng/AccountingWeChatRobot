package logic

import "github.com/eatmoreapple/openwechat"

type taskFn func(msg *openwechat.Message)

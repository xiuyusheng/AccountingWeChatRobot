package bot

import (
	"fmt"

	"github.com/eatmoreapple/openwechat"
)

type MsgHandle func(msg *openwechat.Message) bool

type weChatBot struct {
	*openwechat.Bot
	MsgHandles []MsgHandle
}

func NewWeChatBot(msgH ...MsgHandle) (self *weChatBot) {
	bot := openwechat.DefaultBot(openwechat.Desktop)
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")

	defer reloadStorage.Close()

	// 登陆
	if err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		fmt.Println(err)
		return
	}
	botW := &weChatBot{Bot: bot}
	if msgH != nil {
		botW.AddMsgHandle(msgH...)
	}
	return botW
}

func (wbot *weChatBot) AddMsgHandle(msgH ...MsgHandle) {
	wbot.MsgHandles = append(wbot.MsgHandles, msgH...)
	wbot.Bot.MessageHandler = func(msg *openwechat.Message) {
		for _, v := range wbot.MsgHandles {
			if !v(msg) {
				break
			}
		}
	}
}

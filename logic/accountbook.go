package logic

import (
	"AccountingWeChatRobot/api/gpt"
	"AccountingWeChatRobot/bot"
	"AccountingWeChatRobot/db"
	"AccountingWeChatRobot/db/model"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/eatmoreapple/openwechat"
)

const (
	taskKey   = "taskKey"
	taskValue = "taskValue"
)

var reg *regexp.Regexp = regexp.MustCompile(`@jizhangmingling:{([\d\.]*?)}{(.*?)}{(.*?)}{(.*?)}`)

type confirmTask struct {
	Result chan bool
	Task   string
}

var ConfirmTask chan confirmTask = make(chan confirmTask, 1)

var taskMap = map[string]openwechat.MessageHandler{
	"记账": Bookkeeping,
	"查账": GetFTTable,
}

/* ---------------------------------- 核心业务 ---------------------------------- */
func Bookkeeping(msg *openwechat.Message) {
	if content, ok := msg.Get(taskValue); ok {
		content_, ok := content.(string)
		if ok {
			answer, err := gpt.SysRecord.Ask(fmt.Sprintf("解析出其中金额（阿拉伯数字），菜品种类(多个用‘、’分隔)，食用者名称(多个用‘、’分隔)，未食用者名称(多个用‘、’分隔),并返回格式为‘@jizhangmingling:{金额}{菜品种类}{食用者名称}{未食用者名称}’，例如：菜钱：十一块7毛，菜品有萝卜、白菜，有张三、李四、王五吃了，老六没吃。\n\n返回：@jizhangmingling:{11.7}{白菜、萝卜}{张三、李四、王五}{老六}。根据逻辑识别以下账单文本，：%s", content_))
			if err != nil {
				msg.ReplyText("信息解析错误了")
				return
			}
			if reg.MatchString(answer) {
				ss := reg.FindAllStringSubmatch(answer, 4)[0][1:]
				now := time.Now()
				now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				taskInfo := fmt.Sprintf("日期：%s\n金额：%s元\n菜品种类：%s\n食用者名称：%s（共%d人）", now.Format("2006-01-02"), ss[0], ss[1], ss[2], strings.Count(ss[2], "、")+1)
				msg.ReplyText(fmt.Sprintf("请确认账单是否正确，请回复【是】或【否】：\n%s", taskInfo))
				result := make(chan bool)
				ConfirmTask <- confirmTask{
					Result: result,
					Task:   taskInfo,
				}
				go func() {
					if <-result {
						price, err := strconv.ParseFloat(ss[0], 64)
						if err != nil {
							msg.ReplyText("金额格式错误")
						}
						if err := db.DB.Model(&model.Bill{}).Where("Time = ?", now.Unix()).Save(&model.Bill{
							Time:        model.Time{Time: now},
							Consumption: price,
							Note:        ss[1],
							Consumer:    strings.Split(ss[2], "、"),
						}).Error; err != nil {
							msg.ReplyText("账单写入数据库失败")
						} else {
							msg.ReplyText(getFTTable())
						}
					} else {
						msg.ReplyText("今日的账单已取消，请重新录入")
					}
				}()
			}
		}
	}
}

func GetFTTable(msg *openwechat.Message) {
	msg.ReplyText(getFTTable())
}

/* -------------------------------------------------------------------------- */

/* ---------------------------------- 辅助逻辑 ---------------------------------- */
func getFTTable() string {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := weekday - 1
	monday := now.AddDate(0, 0, -offset)
	var bills []model.Bill
	if err := db.DB.Debug().Where("Time >= ?", monday.Unix()).Find(&bills).Error; err != nil {
		return "账单查询数据库失败"
	} else {
		var temMap = map[string]float64{}
		for _, v := range bills {
			averagePrice := math.Ceil(v.Consumption/float64(len(v.Consumer))*100) / 100
			for _, name := range v.Consumer {
				temMap[name] += averagePrice
			}
		}
		var pjftTable = "本周分摊表\n"
		for name, price := range temMap {
			pjftTable += fmt.Sprintf("   %s:🪙%.2f元🪙\n", name, price)
		}
		return fmt.Sprintf("今日的账单已更新,日期：【%s】\n\n%s", now.Format("2006年01月02"), pjftTable)

	}
}
func Monitor() {
	var groups openwechat.Groups
	bot := bot.NewWeChatBot(func(msg *openwechat.Message) bool {
		if msg.IsSendByGroup() && msg.IsAt() {
			group, _ := msg.Sender()
			group_, _ := group.AsGroup()
			if group_.ID() != groups[0].ID() {
				return false
			}
			msg.Content = strings.ReplaceAll(msg.Content, fmt.Sprintf("@%s", msg.Owner().Self().NickName), "")
			msg.Content = strings.TrimSpace(msg.Content)
			return true
		}
		return false
	}, func(msg *openwechat.Message) bool {
		select {
		case x := <-ConfirmTask:
			switch strings.ToLower(msg.Content) {
			case "ok", "yes", "是的", "是", "没错", "没错的", "确定", "对", "没问题", "对的":
				x.Result <- true
			case "no", "不是的", "否", "不是", "错", "错误", "错的", "取消", "不对", "不算", "不对的":
				x.Result <- false
			default:
				ConfirmTask <- x
				msg.ReplyText(fmt.Sprintf("请先确认账单\n\n%s", x.Task))
			}
		default:
			if msg.Content != "" && msg.Content[0] == '@' && msg.Content[1] != '@' {
				if strings.Contains(msg.Content, ":") {
					valueS := strings.SplitN(msg.Content, ":", 2)
					msg.Set(taskKey, valueS[0][1:])
					msg.Set(taskValue, valueS[1])
					return true
				}
				return false
			}
		}

		return false
	}, func(msg *openwechat.Message) bool {
		if v, ok := msg.Get(taskKey); ok {
			v, ok := v.(string)
			if !ok {
				return false
			}
			if vv, ok := taskMap[v]; ok {
				vv(msg)
				return true
			}
		}
		return false
	})
	self, err := bot.GetCurrentUser()
	if err != nil {
		panic(err)
	}
	groups, _ = self.Groups()

	bot.Block()
}

/* -------------------------------------------------------------------------- */

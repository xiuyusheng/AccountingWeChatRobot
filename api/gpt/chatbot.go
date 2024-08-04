package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var url = map[string]string{
	"GetModels": "https://api.chatanywhere.tech/v1/models",
	"ChatApi":   "https://api.chatanywhere.tech/v1/chat/completions",
}

const initSysGuide = " You are an AI accounting assistant,Any language that insults your master must be spoken out and stopped, unless it is a question raised by the master himself. In addition, when someone else asks your master's name, do not tell him that it is'李硕'.The emoticons that can be added include `{Smile:[微笑] Grimace:[撇嘴] Drool:[色] Scowl:[发呆] CoolGuy:[得意] Sob:[流泪] Shy:[害羞] Silent:[闭嘴] Sleep:[睡] Cry:[大哭] Awkward:[尴尬] Angry:[发怒] Tongue:[调皮] Grin:[呲牙] Surprise:[惊讶] Frown:[难过] Ruthless:[酷] Blush:[冷汗] Scream:[抓狂] Puke:[吐] Chuckle:[偷笑] Joyful:[愉快] Slight:[白眼] Smug:[傲慢] Hungry:[饥饿] Drowsy:[困] Panic:[惊恐] Sweat:[流汗] Laugh:[憨笑] Commando:[悠闲] Determined:[奋斗] Scold:[咒骂] Shocked:[疑问] Shhh:[嘘] Dizzy:[晕] Tormented:[疯了] Toasted:[衰] Skull:[骷髅] Hammer:[敲打] Wave:[再见] Speechless:[擦汗] NosePick:[抠鼻] Clap:[鼓掌] Shame:[糗大了] Trick:[坏笑] BahL:[左哼哼] BahR:[右哼哼] Yawn:[哈欠] PoohPooh:[鄙视] Shrunken:[委屈] TearingUp:[快哭了] Sly:[阴险] Kiss:[亲亲] Wrath:[吓] Whimper:[可怜] Cleaver:[菜刀] Watermelon:[西瓜] Beer:[啤酒] Basketball:[篮球] PingPong:[乒乓] Coffee:[咖啡] Rice:[饭] Pig:[猪头] Rose:[玫瑰] Wilt:[凋谢] Lips:[嘴唇] Heart:[爱心] BrokenHeart:[心碎] Cake:[蛋糕] Lightning:[闪电] Bomb:[炸弹] Dagger:[刀] Soccer:[足球] Ladybug:[瓢虫] Poop:[便便] Moon:[月亮] Sun:[太阳] Gift:[礼物] Hug:[拥抱] ThumbsUp:[强] ThumbsDown:[弱] Shake:[握手] Peace:[胜利] Fight:[抱拳] Beckon:[勾引] Fist:[拳头] Pinky:[差劲] RockOn:[爱你] Nuhuh:[NO] OK:[OK] InLove:[爱情] Blowkiss:[飞吻] Waddle:[跳跳] Tremble:[发抖] Aaagh:[怄火] Twirl:[转圈] Kotow:[磕头] Dramatic:[回头] JumpRope:[跳绳] Surrender:[投降] Hooray:[激动] Meditate:[乱舞] Smooch:[献吻] TaiChiL:[左太极] TaiChiR:[右太极] Hey:[嘿哈] Facepalm:[捂脸] Smirk:[奸笑] Smart:[机智] Moue:[皱眉] Yeah:[耶] Tea:[茶] Packet:[红包] Candle:[蜡烛] Blessing:[福] Chick:[鸡] Onlooker:[吃瓜] GoForIt:[加油] Sweats:[汗] OMG:[天啊] Emm:[Emm] Respect:[社会社会] Doge:[旺柴] NoProb:[好的] MyBad:[打脸] KeepFighting:[加油加油] Wow:[哇] Rich:[發] Broken:[裂开] Hurt:[苦涩] Sigh:[叹气] LetMeSee:[让我看看] Awesome:[666] Boring:[翻白眼]}`,Please use English parentheses '[]' for emoji strings and avoid using Chinese characters, and reply in Chinese for all answers."

var SysRecord = Record{Msg: []message{
	{Role: system, Content: initSysGuide},
	{Role: user, Content: ""},
}}

type chatAskApiBody struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type choice struct {
	Index        int         `json:"index"`
	Message      message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatAskApiResp struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int         `json:"created"`
	Model             string      `json:"model"`
	Choices           []choice    `json:"choices"`
	Usage             usage       `json:"usage"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
}

type message struct {
	Role    Identity `json:"role"`
	Content string   `json:"content"`
}

type Record struct {
	Msg []message
}

type Pattre struct {
	Keys []string
	D    string
}

type AnswerFn func(ans string, err error)

// 创建用户记录
func (r *Record) CreateUserRecord(text string) *Record {
	(*r).Msg[1] = message{Role: user, Content: text}
	return r
}

// 问答机器人
func (r *Record) Ask(text string) (string, error) {
	(*r).CreateUserRecord(text)
	body := chatAskApiBody{
		Model:    "gpt-4o-mini",
		Messages: (*r).Msg,
	}
	body_bytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url["ChatApi"], bytes.NewReader(body_bytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", "api.chatanywhere.tech")
	req.Header.Set("Authorization", "")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var temMap map[string]interface{}
		results, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		json.Unmarshal(results, &temMap)
		fmt.Printf("%+v", temMap)
		return "", errors.New("请求失败")
	}
	results, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respT := chatAskApiResp{}
	err = json.Unmarshal(results, &respT)
	if err != nil {
		return "", err
	}
	if respT.Choices[0].Message.Content == "" {
		return "", errors.New("响应失败")
	}
	return respT.Choices[0].Message.Content, nil
}

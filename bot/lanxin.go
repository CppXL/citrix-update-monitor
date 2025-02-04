package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Chestnuts4/citrix-update-monitor/config"
	"github.com/Chestnuts4/citrix-update-monitor/util"
)

const lanxinBotName = "lanxin bot"

var statusCodeMap = map[int]string{
	0:     "OK",
	10000: "API 服务不可得",
	-1:    "缺少 APP_TOKEN",
	-2:    "缺少 USER_TOKEN",
	-3:    "无效的 APP_TOKEN",
	-4:    "无效的 USER_TOKEN",
	-5:    "API 请求路径错误",
	-6:    "缺少 SYNC_TOKEN",
	-7:    "无效的 SYNC_TOKEN",
	-8:    "缺少 HOOK_TOKEN",
	-9:    "无效的 HOOK_TOKEN",
	-100:  "JSON序列化错误",
	-101:  "JSON反序列化错误",
	10001: "API RPC 服务不可得",
	10002: "AUTH认证失败",
	10003: "无效的请求",
	10004: "API服务 不支持",
	10005: "API服务 无权限 接口权限",
	10006: "API服务 版本检查失败",
	10007: "API服务 数据冲突",
	10008: "API服务 数据不存在",
	40013: "请求参数缺少appid",
	40015: "API 请求参数数目错误",
	40016: "API 请求参数 grant_type 类型错误",
	40017: "API 请求参数 secret 错误",
	40018: "APP Secret错误",
	40019: "appId不匹配",
	40030: "无效的API TOKEN",
	40031: "API 访问中缺少API TOKEN 参数",
	40032: "API 访问中缺少 CODE 参数",
	40033: "API 访问中缺少 uri 参数",
	40035: "无效的USER TOKEN",
	40036: "CODE非法",
	40040: "API 访问缺少userid 参数",
	40059: "不支持该Content-Type",
	40060: "消息为空或格式错",
	40061: "消息URL为空或格式错",
	40062: "消息接收者为空或格式错",
	40070: "消息更新失败",
	45000: "请求参数错误",
	45001: "请求头信息错误",
	59000: "bot 不存在",
	59001: "bot 已停用",
	59002: "安全认证不通过",
}

type LanxinBot struct {
	name       string
	secret     string
	webhook    string
	proxy      string
	httpClient *http.Client
	msgChan    chan string
	errChan    chan error
}

func NewLangxinBot(secret string, webhook string, proxy string) (*LanxinBot, error) {
	client, err := util.BuildClientWithProxy(proxy)
	if err != nil {
		return nil, err
	}
	return &LanxinBot{
		name:       lanxinBotName,
		secret:     secret,
		webhook:    webhook,
		proxy:      proxy,
		httpClient: client,
		msgChan:    make(chan string),
		errChan:    make(chan error),
	}, nil
}

func (b *LanxinBot) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case msg := <-b.msgChan:
				err := b.sendMsg(msg)
				if err != nil {
					b.errChan <- err
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-b.errChan:
				log.Println(err)
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Printf("%s started", b.name)
	return nil
}

type MsgData struct {
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

type lanxinMsgData struct {
	Sign      string  `json:"sign"`
	Timestamp string  `json:"timestamp"`
	MsgType   string  `json:"msgType"`
	MsgData   MsgData `json:"msgData"`
}

func (b *LanxinBot) sendMsgAllWebHook(msg string) error {

	return nil
}

func (b *LanxinBot) sendMsg(msg string) error {
	signature := util.LanxinSign(b.secret)
	timestamp := fmt.Sprintf("%v", time.Now().Unix())
	lxMsg := lanxinMsgData{
		Sign:      signature,
		Timestamp: timestamp,
		MsgType:   "text",
	}
	lxMsg.MsgData.Text.Content = msg
	lxMsgBytes, err := json.Marshal(lxMsg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", b.webhook, bytes.NewBuffer(lxMsgBytes))
	if err != nil {
		// Handle error
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		// Handle error
		return err
	}
	defer resp.Body.Close()

	return nil

}

func (b *LanxinBot) SendMsg(msg *config.Msg) error {
	msgStr := util.FormatMsg(msg)
	b.msgChan <- msgStr
	return nil
}

func (b *LanxinBot) GetBotName() string {
	return lanxinBotName
}

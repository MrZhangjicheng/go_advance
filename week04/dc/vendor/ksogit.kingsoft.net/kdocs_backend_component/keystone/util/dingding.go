package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"ksogit.kingsoft.net/kgo/log"
)

// DingdingHookFormLink 钉钉机器人
type DingdingHookFormLink struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PicURL     string `json:"picUrl"`
}

// DingdingHookForm 钉钉机器人
type DingdingHookForm struct {
	Msgtype string               `json:"msgtype"`
	Link    DingdingHookFormLink `json:"link"`
}

// dingSendRequest 发送钉钉请求的底层接口
func dingSendRequest(method string, url string, payload string) ([]byte, error) {
	var body []byte

	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		return body, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return body, err
	}
	defer res.Body.Close()

	body, _ = ioutil.ReadAll(res.Body)
	return body, nil
}

// SendToDingBot 推送钉钉机器人
func SendToDingBot(reqForm DingdingHookForm) {
	if _, ok := os.LookupEnv("KAE_DINGBOT_URL"); !ok {
		return
	}
	postBody, _ := json.Marshal(reqForm)
	res, err := dingSendRequest("POST", os.Getenv("KAE_DINGBOT_URL"), string(postBody))
	log.Debug("ding res: %s %v", string(res), err)
}

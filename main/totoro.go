package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	APPID          = "wx9c5b1dce69c4b639"
	APPSECRET      = "f7808f9ad2cfa7193110f0e9dda922e6"
	WeatTemplateID = "8azYqSzrl5bHuhuRmwdGHNrtnKbg9YPJffSEpGFJk0U" //天气模板ID，替换成自己的
	WeatherKey     = "ba66f6584af74626afccb83cd8a96b3c"            //和风天气私钥
	LocationId     = "101210112"                                   //拱墅的locationId
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type V struct {
	City struct {
		Value string `json:"value"`
		Color string `json:"color"`
	} `json:"city"`
	Wheather struct {
		Value string `json:"value"`
		Color string `json:"color"`
	} `json:"wheather"`
	Temp struct {
		Value string `json:"value"`
		Color string `json:"color"`
	} `json:"temp"`
	Tip struct {
		Value string `json:"value"`
		Color string `json:"color"`
	} `json:"tip"`
	Sentence struct {
		Value string `json:"value"`
		Color string `json:"color"`
	} `json:"sentence"`
}

type Template struct {
	Touser     string `json:"touser"`
	TemplateId string `json:"template_id"`
	Data       struct {
		V
	} `json:"data"`
}

// 获取微信accesstoken
func getaccesstoken() string {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", APPID, APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}

	return token.AccessToken
}

// 获取关注者列表
func getflist(access_token string) []gjson.Result {
	url := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + access_token + "&next_openid="
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取关注列表失败", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return nil
	}
	flist := gjson.Get(string(body), "data.openid").Array()
	return flist
}

// 发送天气预报
func weather() {
	access_token := getaccesstoken()
	if access_token == "" {
		return
	}
	flist := getflist(access_token)
	if flist == nil {
		return
	}
	for _, v := range flist {
		sendweather(access_token, v.Str)
	}

}

// 发送天气
func sendweather(access_token, openid string) {
	tmp, wea, tip := getweather()
	if tmp == "" || wea == "" || tip == "" {
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?" +
		"access_token=" + access_token
	method := "POST"
	//设置模板消息数据
	v := setVData(tmp, wea, tip)
	template := Template{
		Touser:     openid,
		TemplateId: WeatTemplateID,
		Data:       struct{ V }{V: v},
	}
	//发送模板消息
	bytes, err := json.Marshal(template)
	if err != nil {
		return
	}
	payload := strings.NewReader(string(bytes))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	msg := fmt.Sprintf("successs: %s  + weather is ok", openid)
	fmt.Println(msg)
}
func setVData(tmp, wea, tip string) V {
	sen := getSenc()
	//定义模板数据
	v := V{
		City: struct {
			Value string `json:"value"`
			Color string `json:"color"`
		}{
			Value: "拱墅",
			Color: "#a22fdf",
		},
		Wheather: struct {
			Value string `json:"value"`
			Color string `json:"color"`
		}{
			Value: wea,
			Color: "#ed8fe9",
		},
		Temp: struct {
			Value string `json:"value"`
			Color string `json:"color"`
		}{
			Value: tmp,
			Color: "#bb81c5",
		},
		Tip: struct {
			Value string `json:"value"`
			Color string `json:"color"`
		}{
			Value: tip,
			Color: "#d4a7da",
		},
		Sentence: struct {
			Value string `json:"value"`
			Color string `json:"color"`
		}{
			Value: sen,
			Color: "#8034aa",
		},
	}
	return v
}

// 获取天气
func getweather() (string, string, string) {
	url := fmt.Sprintf("https://devapi.qweather.com/v7/weather/3d?location=%s&key=%s", LocationId, WeatherKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取天气失败", err)
		return "", "", ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", "", ""
	}
	data := gjson.Get(string(body), "daily").Array()
	thisday := data[0].String()
	return weacomb(thisday)
}
func weacomb(thisday string) (string, string, string) {
	tempMax := gjson.Get(thisday, "tempMax").Str     //最高气温
	tempMin := gjson.Get(thisday, "tempMin").Str     //最低气温
	textDay := gjson.Get(thisday, "textDay").Str     //白天天气
	textNight := gjson.Get(thisday, "textNight").Str //傍晚天气
	sunset := gjson.Get(thisday, "sunset").Str       //日落
	moonrise := gjson.Get(thisday, "moonrise").Str   //月升
	moonPhase := gjson.Get(thisday, "moonPhase").Str //月相
	tem := fmt.Sprintf("tempMax: %s℃ -- tempMin: %s℃", tempMax, tempMin)
	wea := fmt.Sprintf("morning: %s -- night: %s", textDay, textNight)
	tip := fmt.Sprintf("日落: %s 月升: %s 月相: %s", sunset, moonrise, moonPhase)
	return tem, wea, tip
}
func getSenc() string {
	url := "https://v1.hitokoto.cn/?encode=json&c=k&c=h&c=j"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取天气失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return ""
	}
	hitokoto := gjson.Get(string(body), "hitokoto").Str
	from := gjson.Get(string(body), "from").Str
	//from_who := gjson.Get(string(body), "from_who").Str
	sen := fmt.Sprintf("%s from:%s", hitokoto, from)
	return sen
}
func main() {
	weather()
}

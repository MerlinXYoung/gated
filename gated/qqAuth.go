package main
import (
	"math/rand"
	"crypto/md5"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"net"
	"time"
	"fmt"
	"log"
)
var QQAuthURL = "http://ysdktest.qq.com/auth/qq_check_token" // MATCH /exported string var.*QQAuthURL.*main unexport/
var QQAppid = "1106662470"
var QQAppkey = "ZQMqEM5I4m5jx68q"

type QQAuth struct {

}

func (self *QQAuth)doAuth(openid string, openkey string)(uint64, error){
	now := time.Now().Unix()
	src := fmt.Sprintf("%s%d", QQAppkey, now)
	sig := md5.Sum([]byte(src))
	url := fmt.Sprintf("%s?timestamp=%d&appid=%s&sig=%x&openid=%s&openkey=%s",
		QQAuthURL, now, QQAppid, sig, openid, openkey)
	log.Println("url:", url)
	var netTransport = &http.Transport{
		Dial:(&net.Dialer{
			Timeout: 10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

	}
	var client = &http.Client{
		Timeout: time.Second* 5,
		Transport: netTransport,

	}

	response, err := client.Get(url)
	if err != nil{
		log.Fatal("http get[", url, "] error:", err)
	}
	defer response.Body.Close()
	if response.StatusCode == 200{
		body, _ := ioutil.ReadAll(response.Body)
		log.Println(string(body))
		type Rsp struct{
			Msg string
			Ret int
		}
		var rsp Rsp
		//it := []interface{}{}
		//json.UnmarshalFromString(string(body), &rsp)
		json.Unmarshal(body, &rsp)
		//log.Printf("%v", rsp)
		log.Printf("msg:%s ret:%d", rsp.Msg, rsp.Ret)
		
	}
	return rand.Uint64(), nil
}
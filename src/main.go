package main

import (
	"fmt"
	"net"
	"net/http"
	"encoding/binary"
	"encoding/json"
	proto "github.com/golang/protobuf/proto"
	"gw_cs"
	"gw_ss"
	"log"
	"time"
	"crypto/md5"
	"io/ioutil"
	"io"
	//goczmq "github.com/zeromq/goczmq"
	zmq "github.com/pebbe/zmq4"
) 
// struct Client
// {
// 	var conn net.Conn
// 	var 
// }
func main()  {
	
	dealer, err := zmq.NewSocket(zmq.DEALER)
	
	if err != nil{
		log.Fatal("Error Dealer:", err)
	}
	defer dealer.Close()
	dealer.Connect("tcp://*:30802")
	
	//addr := new 
	listener, err := net.Listen("tcp", "localhost:30801")
	if err != nil{
		log.Fatal("Error Listen", err.Error())
		return
	}
	defer listener.Close()
	go HandleDealer(dealer)

	for{
		conn, err := listener.Accept()
		if err != nil{
			log.Fatal("Error accpeting", err.Error())
			return
		}
		go HandleConn(conn)
	}

}

func HandleDealer(dealer *zmq.Socket){
	for{
		replyHead, err := dealer.RecvBytes(0)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("dealer received head '%s'", string(replyHead[0]))
		ssHead := &gw_ss.Head{}
		proto.Unmarshal(replyHead, ssHead)
		more, err := dealer.GetRcvmore()
		if err != nil{
			log.Fatal(err)
		}
		if more {
			replyBody, err := dealer.RecvBytes(0)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("dealer received body '%s'", string(replyBody[0]))
			switch ssHead.GetMsgid() {
			case gw_ss.EMsgID_ClientNew:
				ssRsp := &gw_ss.ClientNewReq{}
				proto.Unmarshal(replyBody, ssRsp)
				log.Printf("%v", ssRsp)
			case gw_ss.EMsgID_ClientAuth:
				ssRsp := &gw_ss.ClientAuthReq{}
				proto.Unmarshal(replyBody, ssRsp)
				log.Printf("%v", ssRsp)
			case gw_ss.EMsgID_ClientClose:
				ssRsp := &gw_ss.ClientCloseReq{}
				proto.Unmarshal(replyBody, ssRsp)
				log.Printf("%v", ssRsp)
			case gw_ss.EMsgID_Other:

			default:
				log.Fatal("Invalid msgid:", ssHead.GetMsgid())
			} 
		}
		
		
	}
}



// MATCH /unexported method HandleConn 
func HandleConn(conn net.Conn){
	defer conn.Close()
	lenBuf := make([]byte, 4)
	for{
		len, err := conn.Read(lenBuf)
		if err != nil{
			log.Print("reading pkgLen error:", err.Error())
			if err == io.EOF{
				break;
			}
			
		}
		if len != 4{
			log.Fatal("reading pkgLen error:", err.Error())
			
		}
		pkgLen := binary.BigEndian.Uint32(lenBuf)
		pkg := make([]byte, pkgLen)
		curr := pkg
		var cachedLen uint32 = 0
		for{
			len, err = conn.Read(curr)
			log.Print("reading pkgLen error:", err.Error())
			if err == io.EOF{
				break;
			}
			cachedLen += uint32(len)
			//fmt.Printf("Recv data[%d] cachedLen[%d]\n", len, cachedLen)
			log.Printf("Recv data[%d] cachedLen[%d]\n", len, cachedLen)
			if cachedLen == pkgLen{
				break
			}
		}
		HandlePkg(conn, pkg, cachedLen)
	}
}

func HandlePkg(conn net.Conn, pkg []byte, size uint32){
	headLen := binary.BigEndian.Uint16(pkg[:2])
	log.Println("headLen:", headLen)
	head := &gw_cs.Head{}
	err := proto.Unmarshal(pkg[2:headLen+2], head) 
	if err != nil{
		log.Fatal("unmarshaling head error:", err)
	}
	log.Println("msgid:", head.GetMsgid())
	switch head.GetMsgid(){
	case gw_cs.EMsgID_Auth:
		req := &gw_cs.AuthReq{}
		err = proto.Unmarshal(pkg[2+headLen:], req)
		if err != nil{
			log.Fatal("unmarshaling auth req error:", err)
		}
		HandleAuthReq(conn, head, req)
	case gw_cs.EMsgID_Other:
		HandleOtherReq(conn, head, pkg[2+headLen:], size-2-uint32(headLen))
	default:
		log.Fatal("fuck invalid msgid:", head.GetMsgid())
	}
}
var QQAuthURL = "http://ysdktest.qq.com/auth/qq_check_token" // MATCH /exported string var.*QQAuthURL.*main unexport/
var QQAppid = "1106662470"
var QQAppkey = "ZQMqEM5I4m5jx68q"

// MATCH /unexported methord \.HandleAuthReq
func HandleAuthReq(conn net.Conn, head *gw_cs.Head, req *gw_cs.AuthReq){
	log.Println("HandleAuthReq")
	log.Printf("openid:%s openkey:%s", req.GetOpenid(), req.GetOpenkey())

	now := time.Now().Unix()
	src := fmt.Sprintf("%s%d", QQAppkey, now)
	sig := md5.Sum([]byte(src))
	url := fmt.Sprintf("%s?timestamp=%d&appid=%s&sig=%x&openid=%s&openkey=%s",
		QQAuthURL, now, QQAppid, sig, req.GetOpenid(), req.GetOpenkey())
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
}

func HandleOtherReq(conn net.Conn, head *gw_cs.Head, data []byte, len uint32){
	log.Println("HandleOtherReq")
	log.Printf("data:%s", data)
}


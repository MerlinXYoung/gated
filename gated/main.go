package main

import (
	//"fmt"
	"net"
//"net/http"
	"encoding/binary"
	//"encoding/json"
	proto "github.com/golang/protobuf/proto"
	cs "github.com/MerlinXYoung/gate/cs"
	ss "github.com/MerlinXYoung/gate/ss"
	//gated "github.com/MerlinXYoung/gate/gated"
	"log"
	//"time"
	//"crypto/md5"
	//"io/ioutil"
	//"io"
	//goczmq "github.com/zeromq/goczmq"
	//zmq "github.com/pebbe/zmq4"
) 
// struct Client
// {
// 	var conn net.Conn
// 	var 
// }
var backendMgr *BackendMgr
var clientMgr *ClientMgr
func main()  {
	backendMgr = CreateBackendMgr()
	backend, _ := backendMgr.Create(7, "tcp://127.0.0.1:30802")

	clientMgr =  CreateClientMgr()
	//backend := CreateBackend(7, "tcp://127.0.0.1:30802")
	defer backendMgr.Destory()
	//addr := new 
	listener, err := net.Listen("tcp", "localhost:30801")
	if err != nil{
		log.Fatal("Error Listen", err.Error())
		return
	}
	defer listener.Close()

	go HandleBackend(backend)

	for{
		conn, err := listener.Accept()
		if err != nil{
			log.Fatal("Error accpeting", err.Error())
			return
		}
		go HandleConn(conn, backend)
	}

}

func HandleBackend(backend *Backend){
	for{
		head, bodyBuf, err := backend.RecvHeadMsg()
		if err != nil{
			log.Fatal("backend recv error:", err)
		}
		if bodyBuf != nil {
			log.Printf("backend received body [%d]'%s'", len(bodyBuf))
			switch head.GetMsgid() {
			case ss.EMsgID_ClientNew:
				ssRsp := &ss.ClientNewReq{}
				proto.Unmarshal(bodyBuf, ssRsp)
				log.Printf("%v", ssRsp)
			case ss.EMsgID_ClientAuth:
				ssRsp := &ss.ClientAuthReq{}
				proto.Unmarshal(bodyBuf, ssRsp)
				log.Printf("%v", ssRsp)
			case ss.EMsgID_ClientClose:
				ssRsp := &ss.ClientCloseReq{}
				proto.Unmarshal(bodyBuf, ssRsp)
				log.Printf("%v", ssRsp)
			case ss.EMsgID_Other:
				log.Printf("body:", string(bodyBuf))
			default:
				log.Fatal("Invalid msgid:", head.GetMsgid())
			} 
		}
		
		
	}
}




// MATCH /unexported method HandleConn 
func HandleConn(conn net.Conn, backend *Backend){
	client := clientMgr.Create(backend.id, conn)
	
	defer client.Destroy()
	for{
		pkg, _ :=client.Recv()
		HandlePkg(client, pkg)
	}
	//backend.SendClientClose(client)
	
}

func HandlePkg(client *Client, pkg []byte){
	headLen := binary.BigEndian.Uint16(pkg[:2])
	log.Println("headLen:", headLen)
	head := &cs.Head{}
	err := proto.Unmarshal(pkg[2:headLen+2], head) 
	if err != nil{
		log.Fatal("unmarshaling head error:", err)
	}
	log.Println("msgid:", head.GetMsgid())
	switch head.GetMsgid(){
	case cs.EMsgID_Auth:
		req := &cs.AuthReq{}
		err = proto.Unmarshal(pkg[2+headLen:], req)
		if err != nil{
			log.Fatal("unmarshaling auth req error:", err)
		}
		HandleAuthReq(client, head, req)

	case cs.EMsgID_Other:
		HandleOtherReq(client, head, pkg[2+headLen:], uint32(len(pkg)-2-int(headLen)))
	default:
		log.Fatal("fuck invalid msgid:", head.GetMsgid())
	}
}


// MATCH /unexported methord \.HandleAuthReq
func HandleAuthReq(client *Client, head *cs.Head, req *cs.AuthReq){
	log.Println("HandleAuthReq")
	log.Printf("openid:%s openkey:%s", req.GetOpenid(), req.GetOpenkey())

	auth := &QQAuth{}
	uid, _ := auth.doAuth(req.GetOpenid(), req.GetOpenkey())
	client.uid = uid
	backend := backendMgr.Get(client.bid)
	backend.SendClientAuth(client, req.GetOpenid(), req.GetOpenkey())
	
}

func HandleOtherReq(client *Client, head *cs.Head, data []byte, len uint32){
	log.Println("HandleOtherReq")
	log.Printf("data:%s", data)
	backend := backendMgr.Get(client.bid)
	backend.SendClientOther(client, data)
}


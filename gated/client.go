package main
import (
	"net"
	"io"
	"log"
	"errors"
	"encoding/binary"
)
type Client struct{
	id uint32
	bid uint32 
	conn net.Conn
	uid uint64
	
}

func CreateClient(id uint32, bid uint32, conn net.Conn)(*Client){
	self := new(Client)
	self.id = id
	self.bid = bid
	self.conn = conn
	self.uid = 0
	backendMgr.Get(self.bid).SendClientNew(self)
	return self
}

func (self *Client)Destroy(){
	self.conn.Close()
	backendMgr.Get(self.bid).SendClientClose(self)
}

func (self *Client)Recv()([]byte, error){
	lenBuf := make([]byte, 4)
	len, err := self.conn.Read(lenBuf)
	if err != nil{
		log.Print("reading pkgLen error:", err.Error())
		if err == io.EOF{
			return nil, err;
		}
		
	}
	if len != 4{
		log.Print("reading pkgLen error")
		return nil, errors.New("Read PkgLen error")
	}
	pkgLen := binary.BigEndian.Uint32(lenBuf)
	pkg := make([]byte, pkgLen)
	var cachedLen uint32 = 0
	for{
		len, err = self.conn.Read(pkg[cachedLen:])
		if err != nil{
			log.Print("reading pkg error:", err.Error())
			if err == io.EOF{
				break;
			}
		}
		
		cachedLen += uint32(len)
		//fmt.Printf("Recv data[%d] cachedLen[%d]\n", len, cachedLen)
		log.Printf("Recv data[%d] cachedLen[%d]\n", len, cachedLen)
		if cachedLen == pkgLen{
			break
		}
	}
	return pkg, nil
}

type ClientMgr struct{
	clients map[uint32]*Client
	idAlloc uint32
}

func CreateClientMgr()*ClientMgr{
	self := new(ClientMgr)
	if self == nil{
		return self
	}
	self.idAlloc = 0
	self.clients = make(map[uint32]*Client)
	return self
}

func (self *ClientMgr)Destroy(){
	for id:= range self.clients{
		self.clients[id].Destroy()
		delete(self.clients, id)
	}
}

func (self *ClientMgr)Create(bid uint32, conn net.Conn)*Client{
	self.idAlloc += 1
	return CreateClient(self.idAlloc, bid, conn)
}

func (self *ClientMgr)Get(id uint32)*Client{
	client, ok := self.clients[id]
	if ok {
		return client
	}
	return nil
}




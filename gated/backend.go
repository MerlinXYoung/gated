package main
import(
	"log"
	zmq "github.com/pebbe/zmq4"
	proto "github.com/golang/protobuf/proto"
	ss "github.com/MerlinXYoung/gate/ss"
	"github.com/pkg/errors"
)
type Backend struct{
	id uint32
	sock *zmq.Socket
}

func CreateBackend(id uint32, URL string) *Backend{
	self := &Backend{}
	sock, err := zmq.NewSocket(zmq.DEALER)
	if err != nil{
		log.Fatal("Error zmq NewSocket error:", err)
		//log.Println("Error zmq NewSocket error:", err)
	}
	err = sock.Connect(URL)
	if err != nil{
		log.Fatal("Error zmq sock Connect error:", err)
	}
	self.id = id
	self.sock = sock
	return self
}

func (self *Backend)Destory(){
	self.sock.Close()
}

func (self *Backend)Send(data []byte){
	self.sock.SendBytes(data, 0)
}

func (self *Backend)SendMsg(headData []byte, bodyData []byte){

	log.Printf("head len:%d body len:%d", len(headData), len(bodyData))
	if bodyData == nil || len(bodyData) == 0 {
		self.sock.SendBytes(headData, 0)
	}else{
		self.sock.SendBytes(headData, zmq.SNDMORE)
		self.sock.SendBytes(bodyData, 0)
	}
	
}

func (self *Backend)SendClientNew(client *Client){
	head := &ss.Head{
		ClientId : client.id,
		Msgid : ss.EMsgID_ClientNew,
	}
	req := &ss.ClientNewReq{}
	head.ProtoMessage()
	headBuf, _ := proto.Marshal(head)
	bodyBuf, _ := proto.Marshal(req)
	log.Printf("client[%d] SendCientNew", client.id)
	self.SendMsg(headBuf, bodyBuf)
}

func (self *Backend)SendClientAuth(client *Client, openid string, openkey string){
	head := &ss.Head{
		ClientId : client.id,
		Msgid : ss.EMsgID_ClientAuth,
		Uid : client.uid,
	}
	req := &ss.ClientAuthReq{
		Openid : openid,
		Openkey : openkey,
	}
	headBuf, _ := proto.Marshal(head)
	bodyBuf, _ := proto.Marshal(req)
	log.Printf("client[%d] SendClientAuth", client.id)
	self.SendMsg(headBuf, bodyBuf)
}

func (self *Backend)SendClientClose(client *Client){
	head := &ss.Head{
		ClientId : client.id,
		Msgid : ss.EMsgID_ClientClose,
		Uid : client.uid,
	}
	req := &ss.ClientCloseReq{}
	headBuf, _ := proto.Marshal(head)
	bodyBuf, _ := proto.Marshal(req)
	log.Printf("client[%d] SendClientClose", client.id)
	self.SendMsg(headBuf, bodyBuf)
}

func (self *Backend)SendClientOther(client *Client, body []byte){
	head := &ss.Head{
		ClientId : client.id,
		Msgid : ss.EMsgID_Other,
		Uid : client.uid,
	}

	headBuf, _ := proto.Marshal(head)
	log.Printf("client[%d] SendClientOther", client.id)
	self.SendMsg(headBuf, body)
}

func (self *Backend)RecvMsg()([]byte, []byte, error){
	headBuf, err := self.sock.RecvBytes(0)
	if err != nil {
		log.Print(err)
		return nil, nil, errors.New("Backend: sock recv head error")
	}
	log.Printf("self received head [%d]", len(headBuf))
	more, err := self.sock.GetRcvmore()
	if err != nil{
		log.Print(err)
		return nil,nil, errors.New("Backend: sock get recv more error")
	}
	log.Print("more:", more)
	if !more  {
		return headBuf, nil, nil
	}

	bodyBuf, err := self.sock.RecvBytes(0)
	if err != nil {
		log.Print(err)
		return nil,nil, errors.New("Backend: sock recv body error")
	}
	return headBuf, bodyBuf, nil
		
}

func (self *Backend)RecvHeadMsg()( *ss.Head, []byte, error){
	headBuf, bodyBuf, err := self.RecvMsg()
	if err != nil{
		return nil, nil, err
	}
	head := &ss.Head{}
	err = proto.Unmarshal(headBuf, head)
	if err != nil{
		return nil, nil, err
	}
	return head, bodyBuf, nil

}

type BackendMgr struct{ // MATCH 
	backends map[uint32]*Backend
	list []*Backend
	idx uint64
}

func CreateBackendMgr()(*BackendMgr){
	self := &BackendMgr{}
	self.idx = 0
	self.backends = make(map[uint32]*Backend)
	return self

}
func (self *BackendMgr)Destory(){
	for id:= range self.backends{
		self.backends[id].Destory()
		delete(self.backends, id)
	}
}

func (self *BackendMgr)Create(id uint32, URL string)(*Backend, error){
	backend := CreateBackend(id, URL)
	if backend == nil{
		return nil, errors.New("Create Backend error")
	}
	self.list = append(self.list, backend)
	self.backends[id] = backend
	return backend, nil
}

func (self *BackendMgr)GetRoundRobin()(*Backend, error){
	if len(self.list) == 0{
		return nil, errors.New("Backends empty error")
	} 
	self.idx = self.idx+1
	curr := self.idx % uint64(len(self.list))
	return self.list[curr], nil
}

func (self *BackendMgr)Get(id uint32)(*Backend){
	backend, ok := self.backends[id]
	if ok {
		return backend
	}
	return nil
}

// func (self *Backend)RecvClientAuth()( *ss.Head, *ss.ClientAuthRes, error){
// 	head, bodyBuf, err := self.RecvHeadMsg()
// 	if err != nil{
// 		return nil, nil, err
// 	}
	


// }

// func (self *Backend)RecvClientNew()( *ss.Head, *ss.ClientNewRes, error){

// }

// func (self *Backend)RecvClientClose()( *ss.Head, *ss.ClientCloseRes, error){

// }



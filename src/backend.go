package main
import(
	"log"
	zmq "github.com/pebbe/zmq4"
	proto "github.com/golang/protobuf/proto"
	ss "gw_ss"
)
type Backend struct{
	id uint32
	sock *zmq.Socket
}

func CreateBackend(id uint32, URL string) *Backend{
	self := &Backend{}
	sock, err := zmq.NewSocket(zmq.self)
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
	self.sock.SendBytes(headData, 1)
	self.sock.SendBytes(bodyData, 0)
}

func (self *Backend)RecvMsg()([]byte, []byte, error){
	headBuf, err := self.sock.RecvBytes(0)
	if err != nil {
		log.Print(err)
		return nil,nil, errors.New("Backend: sock recv head error")
	}
	log.Printf("self received head '%s'", string(headBuf))
	more, err := self.sock.GetRcvmore()
	if err != nil{
		log.Print(err)
		return nil,nil, errors.New("Backend: sock get recv more error")
	}
	if !more  {
		return headBuf, nil, nil
	}

	bodyBuf, err := self.RecvBytes(0)
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



package serv
import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
)


type Server struct {
	once sync.Once
	id string
	address string
	sync.Mutex
	// session User list
	users map[string]net.Conn
	
}
// new server  NewServer

func Newserver(id,address string) *Server{
	return newServer(id,address)
}

func newServer(id,address string) *Server {
	return &Server{
		id : id,
		address: address,
		users:make(map[string]net.Conn,100)}
}



//start server
func (s *Server) start() error{
	mux := http.NewServeMux()
	log := logrus.WithFields(logrus.Fields{
		"module":"Server",
		"listen":s.address,
		"id":s.id,
	})

	mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
		//step1 升级
		conn,_,_,err := ws.UpgradeHTTP(r,w)
		if err != nil{
			conn.Close()
			return
		}
		//step2 读取 userId

		user:= r.URL.Query().Get("user")
		if user == ""{
			conn.Close()
			return
		}

		// step3 添加到会话管理中

		old,ok := s.addUser(user,conn)

	})
}

func (s *Server) addUser (user string,conn net.Conn)(net.Conn,bool) {
	s.Lock()
	defer s.Unlock()

}
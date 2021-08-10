package serv

import (
	"errors"
	_ "errors"
	_ "fmt"
	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	once    sync.Once
	id      string
	address string
	sync.Mutex
	// session User list
	users map[string]net.Conn
}

// new server  NewServer

func Newserver(id, address string) *Server {
	return newServer(id, address)
}

func newServer(id, address string) *Server {
	return &Server{
		id:      id,
		address: address,
		users:   make(map[string]net.Conn, 100)}
}

//start server
func (s *Server) start() error {
	mux := http.NewServeMux()
	log := logrus.WithFields(logrus.Fields{
		"module": "Server",
		"listen": s.address,
		"id":     s.id,
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//step1 升级
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			conn.Close()
			return
		}
		//step2 读取 userId

		user := r.URL.Query().Get("user")
		if user == "" {
			conn.Close()
			return
		}

		// step3 添加到会话管理中

		old, ok := s.addUser(user, conn)
		if ok {
			//disconn history connection
			old.Close()
		}
		log.Infof("user %s in", user)

		go func(user string, conn net.Conn) {
			//step4 read the message
			err := s.readloop(user, conn)

			if err != nil {
				log.Error(err)
			}
			conn.Close()
			//step5 disconn del user
			s.delUser(user)
			log.Infof("connection of %s closed", user)
		}(user, conn)
	})
	log.Infoln("started")
	return http.ListenAndServe(s.address, mux)
}

func (s *Server) addUser(user string, conn net.Conn) (net.Conn, bool) {
	s.Lock()
	defer s.Unlock()
	old, ok := s.users[user] //返回旧链接
	s.users[user] = conn     //缓存
	return old, ok
}

func (s *Server) delUser(user string) {
	s.Lock()
	defer s.Unlock()
	delete(s.users, user)
}

//shutdown

func (s *Server) Shutdown() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()
		for _, conn := range s.users {
			conn.Close()
		}
	})
}

//read message

func (s *Server) readloop(user string, conn net.Conn) error {
	for {
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}
		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the conn")
		}
		if frame.Header.Masked {
			ws.Cipher(frame.Payload, frame.Header.Mask, 0)
		}
		// resave the notice message
		if frame.Header.OpCode == ws.OpText {
			go s.handle(user, string(frame.Payload))
		}
	}
}

func (s *Server) handle(user string, message string) {

}

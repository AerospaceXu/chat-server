package server

import (
	"chat-server/user"
	"fmt"
	"net"
	"sync"
)

// Server 服务类
type Server struct {
	ip        string
	port      int
	onlineMap map[string]*user.User
	mapLock   sync.RWMutex
	message   chan string
}

// NewServer 创建并返回实例对象
func NewServer(ip string, port int) *Server {
	return &Server{
		ip:        ip,
		port:      port,
		onlineMap: make(map[string]*user.User),
		message:   make(chan string, 10),
	}
}

// Start 启动服务
func (server *Server) Start() {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", server.ip, server.port),
	)
	if err != nil {
		fmt.Println("net listen error: ", err)
		return
	}
	defer listener.Close()

	go server.listenMessages()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}
		go server.handleUserLogin(connection)
	}
}

// handleUserLogin deal user login
func (server *Server) handleUserLogin(connection net.Conn) {
	user := user.NewUser(connection)
	fmt.Println("登录成功", user.Name)

	server.mapLock.Lock()
	server.onlineMap[user.Name] = user
	server.mapLock.Unlock()

	server.broadMessage(user, "已上线")

	select {}
}

// broadMessage 向当前 user 的 channel 中广播消息
func (server *Server) broadMessage(currentUser *user.User, msg string) {
	currentMsg :=
		"[" + currentUser.Name + "] " + currentUser.Address + ": " + msg
	server.message <- currentMsg
}

// listenMessages listen server.message`s change and send to each user
func (server *Server) listenMessages() {
	for {
		msg, ok := <-server.message
		if ok {
			server.mapLock.Lock()
			for _, user := range server.onlineMap {
				userChannel := user.Channel
				userChannel <- msg
			}
			server.mapLock.Unlock()
		}
	}
}

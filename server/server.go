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

	go server.ListenMessages()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		go server.Handler(connection)
	}
}

// Handler 处理连接
func (server *Server) Handler(connection net.Conn) {
	user := user.NewUser(connection)
	fmt.Println("登录成功", user.GetName())
	userName := user.GetName()
	server.mapLock.Lock()
	server.onlineMap[userName] = user
	server.mapLock.Unlock()

	server.BroadMessage(user, "已上线")

	select {}
}

// BroadMessage 向当前 user 的 channel 中广播消息
func (server *Server) BroadMessage(currentUser *user.User, msg string) {
	userName := currentUser.GetName()
	userAddress := currentUser.GetAddress()

	currentMsg := "[" + userName + "] " + userAddress + " ：" + msg

	server.message <- currentMsg
}

// ListenMessages 监听所有的 message 改变
func (server *Server) ListenMessages() {
	for {
		msg := <-server.message
		server.mapLock.Lock()
		for _, user := range server.onlineMap {
			userChannel := user.GetChannel()
			userChannel <- msg
		}
		server.mapLock.Unlock()
	}
}

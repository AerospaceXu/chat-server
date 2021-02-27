package controller

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Server 服务类
type Server struct {
	ip        string
	port      int
	onlineMap map[string]*User
	mapLock   sync.RWMutex
	message   chan string
}

// NewServer 创建并返回实例对象
func NewServer(ip string, port int) *Server {
	return &Server{
		ip:        ip,
		port:      port,
		onlineMap: make(map[string]*User),
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
	user := NewUser(connection, server)
	fmt.Println("登录成功", user.name)
	user.online()
	isUserAlive := make(chan bool)

	defer func() {
		user.server.mapLock.Lock()
		delete(server.onlineMap, user.name)
		user.server.mapLock.Unlock()
		close(user.channel)
		connection.Close()
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := connection.Read(buf)
			if n == 0 {
				user.offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("connection read error: ", err)
				return
			}

			msg := string(buf[:n-1])
			user.sendMessage(msg)
			isUserAlive <- true
		}
	}()

	for {
		select {
		case <-isUserAlive:
		case <-time.After(time.Second * 100):
			user.showMessage("因长时间未活动，您已被踢出聊天")
			user.sendMessage("因为 [" + user.name + "] 长时间未活动，已被移出聊天室")
			return
		}
	}
}

// broadMessage 向当前 user 的 channel 中广播消息
func (server *Server) broadMessage(currentUser *User, msg string) {
	currentMsg :=
		"[" + currentUser.name + "] " + currentUser.address + ": " + msg
	server.message <- currentMsg
}

// listenMessages listen server.message`s change and send to each user
func (server *Server) listenMessages() {
	for {
		msg, ok := <-server.message
		if ok {
			server.mapLock.Lock()
			for _, user := range server.onlineMap {
				userChannel := user.channel
				userChannel <- msg
			}
			server.mapLock.Unlock()
		}
	}
}

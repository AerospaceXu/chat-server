package controller

import (
	"net"
)

// User struct
type User struct {
	name       string
	address    string
	channel    chan string
	connection net.Conn
	server     *Server
}

// NewUser returns an Users`s instance
func NewUser(connection net.Conn, server *Server) *User {
	userAddress := connection.RemoteAddr().String()
	userName := "name_" + userAddress

	user := &User{
		address:    userAddress,
		name:       userName,
		channel:    make(chan string),
		connection: connection,
		server:     server,
	}

	go user.listenMessage()

	return user
}

// listenMessage receive messsage from other place, then notice user
func (user *User) listenMessage() {
	for {
		msg := <-user.channel

		user.connection.Write([]byte(msg + "\n"))
	}
}

func (user *User) online() {
	user.server.mapLock.Lock()
	user.server.onlineMap[user.name] = user
	user.server.mapLock.Unlock()
	user.server.broadMessage(user, "已上线")
}

func (user *User) offline() {
	user.server.mapLock.Lock()
	delete(user.server.onlineMap, user.name)
	user.server.mapLock.Unlock()
	user.server.broadMessage(user, "已下线")
}

func (user *User) sendMessage(content string) {
	user.server.broadMessage(user, content)
}

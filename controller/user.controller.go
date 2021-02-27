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
		user.showMessage(msg + "\n")
	}
}

// showMessage show message in user`s interface
func (user *User) showMessage(content string) {
	user.connection.Write([]byte(content + "\n"))
}

// online user online system
func (user *User) online() {
	user.server.mapLock.Lock()
	user.server.onlineMap[user.name] = user
	user.server.mapLock.Unlock()
	user.server.broadMessage(user, "已上线")
}

// offline user offline system
func (user *User) offline() {
	user.server.mapLock.Lock()
	delete(user.server.onlineMap, user.name)
	user.server.mapLock.Unlock()
	user.server.broadMessage(user, "已下线")
}

// sendMessage user sendMessage to everybody
func (user *User) sendMessage(content string) {
	if content == "who" {
		onlineUsers := ""
		user.server.mapLock.Lock()
		for _, onlineUser := range user.server.onlineMap {
			onlineUsers += "[" + onlineUser.name + "] 在线\n"
		}
		user.server.mapLock.Unlock()
		user.showMessage(onlineUsers)
	} else if len(content) >= 7 && content[:7] == "rename|" {
		newUserName := content[7:]
		if len(newUserName) < 2 {
			user.showMessage("用户名至少为 2 个字符！！！")
		} else {
			if _, isUserExist := user.server.onlineMap[newUserName]; isUserExist {
				user.showMessage("用户名已被占用，请重新输入！！！")
			} else {
				user.server.mapLock.Lock()
				delete(user.server.onlineMap, user.name)
				user.server.onlineMap[newUserName] = user
				user.server.mapLock.Unlock()
				user.showMessage("修改成功！")
				user.name = newUserName
			}
		}
	} else {
		user.server.broadMessage(user, content)
	}
}

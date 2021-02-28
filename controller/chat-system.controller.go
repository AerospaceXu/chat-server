package controller

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ChatSystem struct {
	onlineUsersLock       sync.RWMutex
	onlineUsers           map[string]*User
	broadMessageChannel   chan string
	consoleMessageChannel chan string
}

func CreateChatSystem(listener net.Listener) {
	chatSystem := &ChatSystem{
		onlineUsers:           make(map[string]*User),
		broadMessageChannel:   make(chan string, 10),
		consoleMessageChannel: make(chan string, 10),
	}

	go chatSystem.listenBroadMessages()
	go chatSystem.listenConsoleMessages()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("connect error: ", err)
			continue
		}

		go chatSystem.handleUserLogin(connection)
	}
}

func (chatSystem *ChatSystem) handleUserLogin(connection net.Conn) {
	defer connection.Close()

	user := chatSystem.handleUserOnline(connection)
	defer close(user.message)

	chatSystem.listenUserActive(user)
}

func (chatSystem *ChatSystem) handleUserOnline(connection net.Conn) *User {
	user := chatSystem.handleUserRegister(connection)
	userLoginConsoleMsg :=
		"[" + user.name + "] 已于 ip: " + user.address + " 上线..."
	userLoginBroadMsg := "[" + user.name + "] 已上线~"
	chatSystem.consoleMessageChannel <- userLoginConsoleMsg
	chatSystem.broadMessageChannel <- userLoginBroadMsg
	return user
}

func (chatSystem *ChatSystem) handleUserOffline(user *User) {
	chatSystem.onlineUsersLock.Lock()
	delete(chatSystem.onlineUsers, user.name)
	chatSystem.onlineUsersLock.Unlock()

	userOfflineMsg := "用户 [" + user.name + "] 已下线"
	chatSystem.consoleMessageChannel <- userOfflineMsg
	chatSystem.broadMessageChannel <- userOfflineMsg
}

func (chatSystem *ChatSystem) handleUserRegister(connection net.Conn) *User {
	userAddress := connection.RemoteAddr().String()
	userName := ""

	chatSystem.onlineUsersLock.Lock()
	for {
		rand.Seed(time.Now().Unix())
		userNameSuffix := strconv.Itoa(rand.Intn(9000) + 1000)
		userName = "用户_" + userNameSuffix
		_, isUserExist := chatSystem.onlineUsers[userName]
		if !isUserExist {
			break
		}
	}
	user := CreateUser(connection, userName, userAddress)
	chatSystem.onlineUsers[userName] = user
	chatSystem.onlineUsersLock.Unlock()

	return user
}

func (chatSystem *ChatSystem) broadMessage(content string) {
	chatSystem.onlineUsersLock.Lock()
	for _, user := range chatSystem.onlineUsers {
		user.showMessage(content)
	}
	chatSystem.onlineUsersLock.Unlock()
}

func (chatSystem *ChatSystem) listenUserActive(user *User) {
	userMessage := make([]byte, 4096)

	isUserAlive := make(chan bool)
	go func() {
		for {
			userMessageLen, err := user.connection.Read(userMessage)

			if userMessageLen == 0 {
				chatSystem.handleUserOffline(user)
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("connection read error: ", err)
				return
			}

			msg := string(userMessage[:userMessageLen-1])
			isUserAlive <- true
			chatSystem.handleUserActive(user, msg)
		}
	}()

	for {
		select {
		case <-isUserAlive:
		case <-time.After(time.Second * 300):
			user.showMessage("因长时间未活动，您已被踢出聊天")
			chatSystem.deleteUser(user.name)
			chatSystem.broadMessageChannel <- "因为 [" + user.name + "] 长时间未活动，已被移出聊天室"
			return
		}
	}
}

func (chatSystem *ChatSystem) handleUserActive(user *User, content string) {
	nowTimeStr := time.Now().Format("2006-01-02 15:04:05")
	userName := "[" + user.name + "]"

	cmdArgsArr := strings.Split(content, "|")

	if isWhoCmd := content == "who"; isWhoCmd {
		onlineUsersStr := chatSystem.getOnlineUsers()
		user.showMessage(onlineUsersStr)
	} else if isRenameCmd := len(content) > 7 && cmdArgsArr[0] == "rename"; isRenameCmd {
		newName := content[7:]
		msg, _ := chatSystem.renameUser(user, newName)
		user.showMessage(msg)
	} else if isToCmd := content[:3] == "to|" && len(cmdArgsArr) >= 3; isToCmd {
		targetUserName := cmdArgsArr[1]
		chatContent := strings.Join(cmdArgsArr[2:], "")
		chatSystem.privateChat(user, targetUserName, chatContent)
	} else {
		userSeeingMsg := userName + nowTimeStr + ": " + content
		chatSystem.broadMessageChannel <- userSeeingMsg
	}
}

func (chatSystem *ChatSystem) privateChat(user *User, targetUserName string, content string) {
	targetUser, isUserExist := chatSystem.findUser(targetUserName)
	if isUserExist {
		targetSeeingMsg := "[" + user.name + "]: " + content
		targetUser.showMessage(targetSeeingMsg)
	} else {
		user.showMessage("用户不存在！")
	}
}

func (chatSystem *ChatSystem) findUser(userName string) (*User, bool) {
	chatSystem.onlineUsersLock.Lock()
	user, isUserExist := chatSystem.onlineUsers[userName]
	chatSystem.onlineUsersLock.Unlock()
	return user, isUserExist
}

func (chatSystem *ChatSystem) deleteUser(userName string) bool {
	_, isUserExist := chatSystem.findUser(userName)
	if isUserExist {
		chatSystem.onlineUsersLock.Lock()
		delete(chatSystem.onlineUsers, userName)
		chatSystem.onlineUsersLock.Unlock()
		return true
	}
	return false
}

func (chatSystem *ChatSystem) renameUser(
	user *User,
	newName string,
) (msg string, success bool) {
	if len(newName) < 2 {
		msg = "用户名必须多于 2 个字符！！！"
		success = false
	} else {
		_, isUserExist := chatSystem.findUser(newName)
		if isUserExist {
			msg = "用户名重复！！！"
			success = false
		} else {
			chatSystem.deleteUser(user.name)
			user.changeName(newName)
			chatSystem.onlineUsers[newName] = user
			msg = "修改成功！！！你好，" + newName + "~"
			success = true
		}
	}
	return
}

func (chatSystem *ChatSystem) getOnlineUsers() (onlineUsersStr string) {
	chatSystem.onlineUsersLock.Lock()
	for _, onlineUser := range chatSystem.onlineUsers {
		onlineUsersStr += "[" + onlineUser.name + "] 在线\n"
	}
	chatSystem.onlineUsersLock.Unlock()
	return
}

func (chatSystem *ChatSystem) listenBroadMessages() {
	for {
		message := <-chatSystem.broadMessageChannel
		chatSystem.broadMessage(message)
	}
}

func (chatSystem *ChatSystem) listenConsoleMessages() {
	for {
		message := <-chatSystem.consoleMessageChannel
		fmt.Println(message)
	}
}

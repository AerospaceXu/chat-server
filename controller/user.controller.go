package controller

import "net"

type User struct {
	name       string
	address    string
	message    chan string
	connection net.Conn
}

func CreateUser(
	connection net.Conn,
	userName string,
	userAddress string,
) *User {
	return &User{
		name:       userName,
		address:    userAddress,
		message:    make(chan string),
		connection: connection,
	}
}

func (user *User) showMessage(content string) {
	user.connection.Write([]byte(content + "\n"))
}

func (user *User) sendMessage(content string) {
	user.message <- content
}

func (user *User) changeName(newName string) {
	user.name = newName
}

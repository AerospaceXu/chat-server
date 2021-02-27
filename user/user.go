package user

import "net"

// User struct
type User struct {
	Name       string
	Address    string
	Channel    chan string
	connection net.Conn
}

// NewUser returns an Users`s instance
func NewUser(connection net.Conn) *User {
	userAddress := connection.RemoteAddr().String()
	userName := "name_" + userAddress

	user := &User{
		Address:    userAddress,
		Name:       userName,
		Channel:    make(chan string),
		connection: connection,
	}

	go user.listenMessage()

	return user
}

// listenMessage receive messsage from other place, then notice user
func (user *User) listenMessage() {
	for {
		msg := <-user.Channel

		user.connection.Write([]byte(msg + "\n"))
	}
}

package user

import "net"

// User struct
type User struct {
	name       string
	address    string
	channel    chan string
	connection net.Conn
}

// NewUser returns an Users`s instance.
func NewUser(connection net.Conn) *User {
	userAddress := connection.RemoteAddr().String()
	userName := "name_" + userAddress

	user := &User{
		address:    userAddress,
		name:       userName,
		channel:    make(chan string),
		connection: connection,
	}

	go user.ListenMessage()

	return user
}

// GetName return instance`s name prop value.
func (user *User) GetName() string {
	return user.name
}

// GetAddress return instance`s address prop value.
func (user *User) GetAddress() string {
	return user.address
}

// GetChannel return instance`s channel prop value.
func (user *User) GetChannel() chan string {
	return user.channel
}

// ListenMessage listend messsage transfer.
func (user *User) ListenMessage() {
	for {
		msg := <-user.channel

		user.connection.Write([]byte(msg + "\n"))
	}
}

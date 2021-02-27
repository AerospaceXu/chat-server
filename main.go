package main

import "chat-server/server"

func main() {
	app := server.NewServer("192.168.31.249", 8000)
	app.Start()
}

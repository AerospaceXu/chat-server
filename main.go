package main

import "chat-server/server"

func main() {
	app := server.NewServer("localhost", 8000)
	app.Start()
}

package main

import "chat-server/server"

func main() {
	appServer := server.CreateServer("localhost", 8000)
	appServer.Start()
}

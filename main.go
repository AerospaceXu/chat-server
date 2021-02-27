package main

import "chat-server/controller"

func main() {
	app := controller.NewServer("192.168.31.249", 8000)
	app.Start()
}

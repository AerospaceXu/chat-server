package server

import (
	"chat-server/controller"
	"fmt"
	"net"
)

type Server struct {
	ip   string
	port int
}

func CreateServer(ip string, port int) *Server {
	return &Server{
		ip:   ip,
		port: port,
	}
}

func (server *Server) Start() {
	urlAddress := fmt.Sprintf("%s:%d", server.ip, server.port)
	listener, err := net.Listen("tcp", urlAddress)
	if err != nil {
		fmt.Println("tcp listen err: ", err)
		return
	}
	defer listener.Close()
	fmt.Println("服务器启动成功")

	controller.CreateChatSystem(listener)
}

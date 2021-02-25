package server

import (
	"fmt"
	"net"
)

// Server 服务类
type Server struct {
	IP   string
	Port int
}

// NewServer 创建并返回实例对象
func NewServer(ip string, port int) *Server {
	return &Server{ip, port}
}

// Start 启动服务
func (server *Server) Start() {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", server.IP, server.Port),
	)
	if err != nil {
		fmt.Println("net listen error: ", err)
		return
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		go server.Handler(connection)
	}
}

// Handler 处理连接
func (server *Server) Handler(connection net.Conn) {
	fmt.Println("建立连接成功")
}

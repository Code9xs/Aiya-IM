package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播的channel
	Message chan string
}

// NewServer 创建一个Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// BroadCast 广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	// ... 当前连接的业务
	//fmt.Println("连接建立成功")
	user := NewUser(conn, server)
	// 用户上线，将用户加入到onlineMap中
	user.Online(conn)
	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			// n: 消息的长度
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])
			// 得到的消息进行广播
			user.DoMessage(msg)
		}
	}()
	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，此case不再阻塞，进行下一次循环
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 900):
			// 超出等待时间没有进入下一次循环，已经超时
			// 将当前User强制关闭
			user.SendMsg("You got kicked out of chat for timeout")
			close(user.C)
			conn.Close()

			// 退出当前Handler
			return
		}
	}
}

// ListenMessager 箭筒Message广播消息的channel的goroutine，一旦有消息就发送给全部的在线User
func (server *Server) ListenMessager() {
	for {
		msg := <-server.Message
		// 将msg发送给全部的在线User
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// Start 启动服务的接口
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("new.Listen err: ", err)
		return
	}
	// close listen socket
	defer listener.Close()
	// 启动监听Message的goroutine
	go server.ListenMessager()
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
			continue
		}
		// do handler
		go server.Handler(conn)
	}

}

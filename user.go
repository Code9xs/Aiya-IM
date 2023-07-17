package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

// ListenMessage 监听当前 User channel的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}

// Online 用户的上线业务
func (user *User) Online(conn net.Conn) {
	// 用户上线，将用户加入到onlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	// 广播当前用户上线消息
	user.server.BroadCast(user, "online")
}

// Offline 用户的下线业务
func (user *User) Offline() {
	// 用户下线 将用户从onlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	// 广播当前用户上线消息
	user.server.BroadCast(user, "offline")
}

// SendMsg 给当前用户发消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

// DoMessage 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	if msg == "all" {
		// 查询当前在线用户都有哪些
		user.server.mapLock.Lock()
		for _, iUser := range user.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + iUser.Name + ":" + "Online...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|Jim
		newName := strings.Split(msg, "|")[1]
		// 判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("Current username is exist!")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("Rename success! New name is " + newName + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式： to|Jim|消息内容
		msgArr := strings.Split(msg, "|")
		// 获取对方的用户名
		targetUserName := msgArr[1]
		// 根据用户名 得到对方的User对象
		targetUser, ok := user.server.OnlineMap[targetUserName]
		if !ok {
			// 判断目标用户不存在的情况
			user.SendMsg("User " + targetUserName + " not exist!\n")
		} else {
			// 获取消息内容，通过对方的User对象将消息内容发送过去
			content := msgArr[2]
			// 内容为空
			if content == "" {
				user.SendMsg("No content, please resend!")
				return
			}
			targetUser.SendMsg("[" + user.Addr + "]" + user.Name + " to you:" + content + "\n")
		}
	} else {
		user.server.BroadCast(user, msg)
	}
}

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	mode       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		mode:       999,
	}
	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	// 返回对象
	client.conn = conn
	return client
}

func (client *Client) menu() bool {
	var inputVal int
	fmt.Println("1.Public chat")
	fmt.Println("2.Private chat")
	fmt.Println("3.Update username")
	fmt.Println("4.Show all online user")
	fmt.Println("0.exit")
	fmt.Scanln(&inputVal)
	if inputVal >= 0 && inputVal <= 4 {
		client.mode = inputVal
		return true
	} else {
		fmt.Println(">>>>Please enter a legal option(0 - 4)<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.mode != 0 {
		for client.menu() != true {
			fmt.Println(">>>> Client closed")
		}
		// 根据不同的模式处理不同的业务
		switch client.mode {
		case 1:
			// 公聊
			client.PublicChat()
			break
		case 2:
			// 私聊
			client.PrivateChat()
			break
		case 3:
			// 修改用户名
			client.UpdateName()
			break
		case 4:
			// 查看所有在线用户
			client.ShowAllUser()
			break
		}
	}
}

func (client *Client) UpdateName() bool {
	fmt.Printf(">>>>Enter your new username:\n")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (client *Client) ShowAllUser() bool {
	sendMsg := "all\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>> Entry your message, enter 'exit' to exit")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error: ", err)
				break
			}
			chatMsg = ""
			fmt.Println(">>>> Entry your message, enter 'exit' to exit")
			fmt.Scanln(&chatMsg)
		}
	}
}

func (client *Client) PrivateChat() {
	var targetName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>>> Please enter the chat partner name, enter 'exit' to exit")
	fmt.Scanln(&targetName)

	for targetName != "exit" {
		fmt.Println(">>>> Please enter the message, enter 'exit' to exit")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + targetName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error: ", err)
					break
				}
				chatMsg = ""
				fmt.Println(">>>> Entry your message, enter 'exit' to exit")
				fmt.Scanln(&chatMsg)
			}
		}
		client.SelectUsers()
		fmt.Println(">>>> Please enter the chat partner name, enter 'exit' to exit")
		fmt.Scanln(&targetName)
	}
}

func (client *Client) SelectUsers() {
	sendMsg := "all\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return
	}
}

// DealResponse 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 一旦client.conn 有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

var serverIp string
var serverPort int

func init() {
	// ./client -ip 127.0.0.1 -port 8888
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server ip address(default 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "set server port(default 8888)")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> link server failed")
		return
	}
	// 单独开启一个 goroutine 处理server的回执消息
	go client.DealResponse()
	fmt.Println(">>>>>> link server succeed")
	// 启动客户端业务
	client.Run()
}

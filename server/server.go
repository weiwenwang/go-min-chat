package main

import (
	"net"
	"fmt"
	"os"
	"go-min-chat/server/ser"
	"go-min-chat/mysql"
	"go-min-chat/Utils"
	"go-min-chat/Msg"
	"flag"
)

func init() {
	MinChatSer := ser.GetMinChatSer()
	ini_parser := Utils.IniParser{}
	conf_file_name := "conf.ini"
	if err := ini_parser.Load("../conf/conf.ini"); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", conf_file_name, err.Error())
		return
	}
	MinChatSer.Host = ini_parser.GetString("test", "ip")
	MinChatSer.Port = int(ini_parser.GetInt32("test", "port"))

	flag.StringVar(&MinChatSer.Host, "h", MinChatSer.Host, "is port")

	flag.IntVar(&MinChatSer.Port, "p", MinChatSer.Port, "is port")
	flag.Parse()
}

func main() {
	MinChatSer := ser.GetMinChatSer()
	addr := fmt.Sprintf("%s:%d", MinChatSer.Host, MinChatSer.Port)
	fmt.Println(addr)
	listen, err := net.Listen("tcp", addr)
	Utils.CheckError(err)
	defer listen.Close()
	fmt.Println("Ready to accept connections")
	var u *mysql.User
	for {
		newConn, err := listen.Accept()
		// 连接上了，就把这个连接赋予一个未登录的用户
		u = mysql.BuildUser(0, "", 0, false)
		MinChatSer.AllUser[newConn] = u
		fmt.Println(newConn.RemoteAddr())
		Utils.CheckError(err)
		ch := make(chan []byte)
		go recvConnMsg(newConn, ch)
		go sendConnMsg(newConn, ch)
	}
}

// 服务端接受消息
func recvConnMsg(conn net.Conn, ch chan []byte) {
	buf := make([]byte, 50)
Loop:
	for {
		n, err := conn.Read(buf)
		ret := Utils.ConnReadCheckError(err, conn)
		if (ret == 0) { // 读取时, 发生了错误
			os.Exit(1)
		} else if (ret == -1) { // 客户端断开了连接
			break Loop
		}
		ch <- buf[:n]
	}
}

// 服务端发送消息
func sendConnMsg(conn net.Conn, ch chan []byte) {
	for {
		content, _ := <-ch
		Msg.DoAllMsg(conn, content)
	}
}

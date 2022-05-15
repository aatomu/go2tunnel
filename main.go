package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

var (
	UseProtcol                    = "tcp4"
	ServerLocalAddress            = "localhost:25564"
	ProxyGlobalAddress            = "atomic.f5.si:25200"
	ProxyListen                   = ":25200"
	ClientListen                  = ":25300"
	BootServer         ServerType = Server
)

type ServerType int8

var (
	Server ServerType = 1
	Proxy  ServerType = 2
	//Client ServerType = 3
)

func main() {
	// 鯖ごとに分岐
	switch BootServer {
	case Server:
		// 複数 Session 生成できるように Loop
		for {
			// ProxyとのSesison作成
			PrintInfo("Dial Up to Proxy")
			proxy, err := net.Dial(UseProtcol, ProxyGlobalAddress)
			ErrorCheck("Proxy", err)
			// ServerとのSesison作成
			PrintInfo("Dial Up to Server")
			server, err := net.Dial(UseProtcol, ServerLocalAddress)
			ErrorCheck("Server", err)
			// Sessionが使われるまで待機
			for {
				buf := make([]byte, 128)
				n, _ := proxy.Read(buf)
				if string(buf[:n]) == "Next" {
					break
				}
			}
			// Proxy Sesison <=> Server Sesison を接続
			go copyIO(proxy, server)
			go copyIO(server, proxy)
		}
	case Proxy:
		// ServerからのSesison Trigger 作成
		server, err := net.Listen(UseProtcol, ProxyListen)
		ErrorCheck("Server", err)
		PrintInfo("Listen Server Session Request")
		// ClientからのSession Trigger 作成
		client, err := net.Listen(UseProtcol, ClientListen)
		ErrorCheck("Client", err)
		PrintInfo("Listen Client Session Request")
		// 複数 Session 生成できるように Loop
		for {
			// Server との Session を待機
			serverConn, err := server.Accept()
			ErrorCheck("Server", err)
			PrintInfo("Connected Server")
			// Client との Session を待機
			clientConn, err := client.Accept()
			ErrorCheck("Client", err)
			PrintInfo("Connected Client")
			// Session を使ったことを通知
			serverConn.Write([]byte("Next"))
			PrintInfo("Sended Use Session Info To Server")
			time.Sleep(1 * time.Second)
			// Client Session <=> Server Session を接続
			go copyIO(serverConn, clientConn)
			go copyIO(clientConn, serverConn)
		}
	}
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}

func ErrorCheck(host string, err error) {
	if err != nil {
		fmt.Printf("[ERROR]: <%s> %s\n", host, err.Error())
	}
}

func PrintInfo(message string) {
	fmt.Printf("[INFO]: %s\n", message)
}

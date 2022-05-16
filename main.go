package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"
)

type Settings struct {
	UseProtcol         string `json:"UseProtcol"`
	ServerLocalAddress string `json:"ServerLocalAddress"`
	ProxyGlobalAddress string `json:"ProxyGlobalAddress"`
	ProxyListen        string `json:"ProxyListen"`
	ClientListen       string `json:"ClientListen"`
	BootServer         string `json:"BootServer"`
}

func main() {
	path := flag.String("env", "./example.json", "flags")
	flag.Parse()
	byteArray, _ := ioutil.ReadFile(*path)
	var settings Settings
	json.Unmarshal(byteArray, &settings)
	// 鯖ごとに分岐
	switch settings.BootServer {
	case "Server":
		// 複数 Session 生成できるように Loop
		for {
			// ProxyとのSesison作成
			PrintInfo(fmt.Sprintf("Dial Up to Proxy: \"%s\"", settings.ProxyGlobalAddress))
			proxyConn, err := net.Dial(settings.UseProtcol, settings.ProxyGlobalAddress)
			ErrorCheck("Proxy", err)
			// ServerとのSesison作成
			PrintInfo(fmt.Sprintf("Dial Up to Server: \"%s\"", settings.ServerLocalAddress))
			serverConn, err := net.Dial(settings.UseProtcol, settings.ServerLocalAddress)
			ErrorCheck("Server", err)
			// Sessionが使われるまで待機
			for {
				buf := make([]byte, 128)
				n, err := proxyConn.Read(buf)
				if string(buf[:n]) == "Next" || err == io.EOF {
					break
				}
				if err != nil {
					ErrorCheck("Proxy failed", err)
				}
			}
			// Proxy Sesison <=> Server Sesison を接続
			PrintInfo(fmt.Sprintf("Connect Session %s <=> %s", proxyConn.RemoteAddr(), serverConn.RemoteAddr()))
			go copyIO(proxyConn, serverConn)
			go copyIO(serverConn, proxyConn)
		}
	case "Proxy":
		// ServerからのSesison Trigger 作成
		server, err := net.Listen(settings.UseProtcol, settings.ProxyListen)
		ErrorCheck("Server", err)
		PrintInfo(fmt.Sprintf("Listen Server Session: \"%s\"", settings.ProxyListen))
		// ClientからのSession Trigger 作成
		client, err := net.Listen(settings.UseProtcol, settings.ClientListen)
		ErrorCheck("Client", err)
		PrintInfo(fmt.Sprintf("Listen Client Session: \"%s\"", settings.ClientListen))
		// 複数 Session 生成できるように Loop
		for {
			// Server との Session を待機
			serverConn, err := server.Accept()
			ErrorCheck("Server", err)
			PrintInfo(fmt.Sprintf("Connected Server: \"%s\"", serverConn.RemoteAddr()))
			// Client との Session を待機
			clientConn, err := client.Accept()
			ErrorCheck("Client", err)
			PrintInfo(fmt.Sprintf("Connected Client: \"%s\"", clientConn.RemoteAddr()))
			// Session を使ったことを通知
			serverConn.Write([]byte("Next"))
			PrintInfo("Request New Session From Server")
			time.Sleep(1 * time.Second)
			// Client Session <=> Server Session を接続
			PrintInfo(fmt.Sprintf("Connect Session %s <=> %s", serverConn.RemoteAddr(), clientConn.RemoteAddr()))
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

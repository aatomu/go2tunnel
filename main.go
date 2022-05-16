package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

var (
	path     = flag.String("env", "./example.json", "flags")
	settings Settings
)

func main() {
	flag.Parse()
	byteArray, _ := ioutil.ReadFile(*path)
	json.Unmarshal(byteArray, &settings)
	// 鯖ごとに分岐
	switch settings.BootServer {
	case "Server":
		var err error
		// 複数 Session 生成できるように Loop
		for {
			var proxyConn, serverConn net.Conn
			// ProxyとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up to Proxy: \"%s\"", settings.ProxyGlobalAddress))
				proxyConn, err = net.Dial(settings.UseProtcol, settings.ProxyGlobalAddress)
				if !isError("Proxy", err) {
					break
				}
				time.Sleep(1 * time.Second)
			}
			// ServerとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up to Server: \"%s\"", settings.ServerLocalAddress))
				serverConn, err = net.Dial(settings.UseProtcol, settings.ServerLocalAddress)
				if !isError("Server", err) {
					break
				}
				time.Sleep(1 * time.Second)
			}
			// Sessionが使われるまで待機
			for {
				buf := make([]byte, 128)
				n, err := proxyConn.Read(buf)
				if string(buf[:n]) == "Next" || err == io.EOF {
					break
				}
				if err != nil {
					isError("Proxy failed", err)
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
		isError("Server", err)
		PrintInfo(fmt.Sprintf("Listen Server Session: \"%s\"", settings.ProxyListen))
		// ClientからのSession Trigger 作成
		client, err := net.Listen(settings.UseProtcol, settings.ClientListen)
		isError("Client", err)
		PrintInfo(fmt.Sprintf("Listen Client Session: \"%s\"", settings.ClientListen))
		// 複数 Session 生成できるように Loop
		for {
			// Server との Session を待機
			serverConn, err := server.Accept()
			isError("Server", err)
			PrintInfo(fmt.Sprintf("Connected Server: \"%s\"", serverConn.RemoteAddr()))
			// Client との Session を待機
			clientConn, err := client.Accept()
			isError("Client", err)
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
	defer func() {
		err := src.Close()
		if err != nil {
			PrintInfo(fmt.Sprintf("Session Closed: \"%s\"", src.RemoteAddr()))
		}
		err = dest.Close()
		if err != nil {
			PrintInfo(fmt.Sprintf("Session Closed: \"%s\"", src.RemoteAddr()))
		}
	}()
	io.Copy(src, dest)
}

func isError(host string, err error) (ok bool) {
	if err != nil {
		log.Printf("[ERROR]: <%s> %s\n", host, err.Error())
		return true
	}
	return false
}

func PrintInfo(message string) {
	log.Printf("[INFO]: %s\n", message)
}

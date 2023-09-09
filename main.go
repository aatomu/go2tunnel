package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Settings struct {
	ToServer       string `json:"ToServer"`
	DialupToProxy  string `json:"DialupToProxy"`
	ListenByServer string `json:"ListenByServer"`
	ListenByClient string `json:"ListenByClient"`
	BootType       string `json:"BootType"`
}

var (
	path     = flag.String("env", "./example.json", "flags")
	settings Settings
	dialer   = net.Dialer{Timeout: 30 * time.Second}
)

func main() {
	flag.Parse()
	byteArray, _ := os.ReadFile(*path)
	json.Unmarshal(byteArray, &settings)
	// 鯖ごとに分岐
	switch settings.BootType {
	case "Server":
		if settings.DialupToProxy == "" {
			panic("Failed Get \"DialupToProxy\" Value")
		}
		if settings.ToServer == "" {
			panic("Failed Get \"ToServer\" Value")
		}

		var err error
		// 複数 Session 生成できるように Loop
		for {
			var proxyConn, serverConn net.Conn
			// ProxyとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up Proxy: \"%s\"", settings.DialupToProxy))
				proxyConn, err = dialer.Dial("tcp", settings.DialupToProxy)
				if !isError("Proxy", err) {
					break
				}
				time.Sleep(5 * time.Second)
			}

			// ServerとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up Server: \"%s\"", settings.ToServer))
				serverConn, err = dialer.Dial("tcp", settings.ToServer)
				if !isError("Server", err) {
					break
				}
				time.Sleep(5 * time.Second)
			}

			// Sessionが使われるまで待機
			for {
				buf := make([]byte, 8)
				n, err := proxyConn.Read(buf)
				if string(buf[:n]) == "Next" || err == io.EOF {
					break
				}
				if err != nil {
					isError("Proxy failed", err)
				}
			}

			// Proxy Sesison <=> Server Sesison を接続
			PrintInfo(fmt.Sprintf("Connect [Proxy]%s <=> [Server]%s", proxyConn.RemoteAddr(), serverConn.RemoteAddr()))
			go copyIO(proxyConn, serverConn, false)
			go copyIO(serverConn, proxyConn, true)
		}

	case "Proxy":
		if settings.ListenByServer == "" {
			panic("Failed Get \"ListenByServer\" Value")
		}
		if settings.ListenByClient == "" {
			panic("Failed Get \"ListenByClient\" Value")
		}

		// ClientからのSession Trigger 作成
		client, err := net.Listen("tcp", settings.ListenByClient)
		isError("Client", err)
		PrintInfo(fmt.Sprintf("Listen Client Session: \"%s\"", settings.ListenByClient))
		// ServerからのSesison Trigger 作成
		server, err := net.Listen("tcp", settings.ListenByServer)
		isError("Server", err)
		PrintInfo(fmt.Sprintf("Listen Server Session: \"%s\"", settings.ListenByServer))

		// 複数 Session 生成できるように Loop
		for {
			// Server との Session を待機
			serverConn, err := server.Accept()
			isError("Server", err)
			PrintInfo(fmt.Sprintf("Catch Server: \"%s\"", serverConn.RemoteAddr()))
			// Client との Session を待機
			clientConn, err := client.Accept()
			isError("Client", err)
			PrintInfo(fmt.Sprintf("Catch Client: \"%s\"", clientConn.RemoteAddr()))

			// Session を使ったことを通知
			serverConn.Write([]byte("Next"))
			time.Sleep(1 * time.Second)

			// Client Session <=> Server Session を接続
			PrintInfo(fmt.Sprintf("[Server]%s <=> [Client]%s (https://ipinfo.io/%s) ", serverConn.RemoteAddr(), clientConn.RemoteAddr(), strings.Split(clientConn.RemoteAddr().String(), ":")[0]))
			// Server <=> Client
			go copyIO(serverConn, clientConn, false)
			go copyIO(clientConn, serverConn, true)
		}
	}
}

func copyIO(src, dest net.Conn, shouldClose bool) {
	defer func() {
		if shouldClose {
			time.Sleep(10 * time.Second)
			PrintInfo(fmt.Sprintf("Close Session %s <=> %s", src.RemoteAddr(), dest.RemoteAddr()))
			err := src.Close()
			isError("IOcopy", err)
			err = dest.Close()
			isError("IOcopy", err)
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

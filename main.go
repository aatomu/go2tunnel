package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

type Settings struct {
	TransferProtcol string `json:"TransferProtcol"`
	ToServer        string `json:"ToServer"`
	DialupToProxy   string `json:"DialupToProxy"`
	ListenByServer  string `json:"ListenByServer"`
	ListenByClient  string `json:"ListenByClient"`
	BootType        string `json:"BootType"`
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
	switch settings.BootType {
	case "Server":
		var err error
		// 複数 Session 生成できるように Loop
		for {
			var proxyConn, serverConn net.Conn
			// ProxyとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up to Proxy: \"%s\"", settings.DialupToProxy))
				proxyConn, err = net.Dial("tcp", settings.DialupToProxy)
				if !isError("Proxy", err) {
					break
				}
				time.Sleep(1 * time.Second)
			}
			// ServerとのSesison作成
			for {
				PrintInfo(fmt.Sprintf("Dial Up to Server: \"%s\"", settings.ToServer))
				serverConn, err = net.Dial(settings.TransferProtcol, settings.ToServer)
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
			PrintInfo(fmt.Sprintf("Connect Session %s <=> %s", proxyConn.RemoteAddr(), serverConn.RemoteAddr()))
			go copyIO(proxyConn, serverConn)
			go copyIO(serverConn, proxyConn)
		}
	case "Proxy":
		// ServerからのSesison Trigger 作成
		server, err := net.Listen("tcp", settings.ListenByServer)
		isError("Server", err)
		PrintInfo(fmt.Sprintf("Listen Server Session: \"%s\"", settings.ListenByServer))
		// Server との Session を待機
		serverConn, err := server.Accept()
		isError("Server", err)
		PrintInfo(fmt.Sprintf("Connected Server: \"%s\"", serverConn.RemoteAddr()))
		// ClientからのSession Trigger 作成
		switch settings.TransferProtcol {
		case "tcp":
			client, err := net.Listen("tcp", settings.ListenByServer)
			isError("Client", err)
			PrintInfo(fmt.Sprintf("Listen Client Session: \"%s\"", settings.ListenByClient))
			// 複数 Session 生成できるように Loop
			for {
				// Client との Session を待機
				clientConn, err := client.Accept()
				isError("Client", err)
				PrintInfo(fmt.Sprintf("Connected Client: \"%s\"", clientConn.RemoteAddr()))
				// Session を使ったことを通知
				serverConn.Write([]byte("Next"))
				PrintInfo("Request New Session From Server")
				time.Sleep(1 * time.Second)
				// Client Session <=> Server Session を接続
				PrintInfo(fmt.Sprintf("Connect Session %s <=> %s (https://ipinfo.io/%s) ", serverConn.RemoteAddr(), clientConn.RemoteAddr(), strings.Split(clientConn.RemoteAddr().String(), ":")[0]))
				// Server <=> Client
				go copyIO(serverConn, clientConn)
				go copyIO(clientConn, serverConn)
			}
		case "udp":
			client, err := net.ListenPacket("udp", settings.ListenByClient)
			isError("Client", err)
			PrintInfo(fmt.Sprintf("Listen Client Session: \"%s\"", settings.ListenByClient))
			// 複数 Session 生成できるように Loop
			for {
				b := make([]byte, 1024)
				// Client との Session を待機
				n, addr, err := client.ReadFrom(b)
				isError("Client", err)
				PrintInfo(fmt.Sprintf("Connected Client: \"%s\"", addr.String()))
				// Session を使ったことを通知
				serverConn.Write([]byte("Next"))
				PrintInfo("Request New Session From Server")
				time.Sleep(1 * time.Second)
				// Client Session <=> Server Session を接続
				PrintInfo(fmt.Sprintf("Connect Session %s <=> %s (https://ipinfo.io/%s) ", serverConn.RemoteAddr(), addr.String(), strings.Split(addr.String(), ":")[0]))
				go func(bytes []byte, n int, addr net.Addr) {
					for {
						serverConn.Write(bytes[:n])
						n, _ := serverConn.Read(bytes)
						client.WriteTo(bytes[:n], addr)
						n, addr, _ = client.ReadFrom(bytes[:n])
					}
				}(b, n, addr)
			}
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
			PrintInfo(fmt.Sprintf("Session Closed: \"%s\"", dest.RemoteAddr()))
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

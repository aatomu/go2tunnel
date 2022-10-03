# go2tunnel
開発環境 : go version go1.17.8 linux/arm64 raspi4 8GB  

## how to use
1. go run main.go -env="???.json" で "BootServer":"Proxy" で起動  
2. go run main.go -env="???.json" で "BootServer":"Server" で起動  
Server Proxy どっちを先に起動しても問題なし

## 変数 について
```json
{
  "comment":"転送するProtcol tcp/udpのみ",
  "UseProtcol":"tcp", 
  "comment":"サーバーへアクセスする際のアドレス",
  "ServerLocalAddress":"localhost:25565",
  "comment":"Server=>ProxyのProxy Address",
  "ProxyGlobalAddress":"example.com:80",
  "comment":"Server=>ProxyのProxy Port",
  "ProxyListen":":80",
  "comment":"Client=>ProxyのProxy Port",
  "ClientListen":":22",
  "comment":"ServerなのかProxyなのかを指定",
  "BootServer": "Server"
}
```
## 動作
```mermaid
  sequenceDiagram
    participant Client
    participant Proxy
    participant Server
    Server->>Proxy: TCP request
    Client->>Proxy: TCP/UDP request
    loop I/O to I/O
      Proxy->>Server: Transfar Data (overTCP)
      Server->>Proxy: 
      Proxy->>Client: 
      Client->>Proxy: 
    end
```

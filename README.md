# go2tunnel
開発環境 : go version go1.17.8 linux/arm64 raspi4 8GB  

## how to use
1. go run main.go -env="???.json" で "BootServer":"Proxy" で起動  
2. go run main.go -env="???.json" で "BootServer":"Server" で起動  
Server Proxy どっちを先に起動しても問題なし

## 変数 について
```json
{
  "comment": "転送するProtcol tcp/tcp4/tcp6のみ",
  "TransferProtcol": "tcp", 
  "comment": "サーバーへアクセスする際のアドレス",
  "ToServer": "localhost:25565",
  "comment": "Server=>ProxyのProxy Address",
  "DialupToProxy": "example.com:80",
  "comment": "Server=>ProxyのProxy Port",
  "ListenByServer": ":80",
  "comment": "Client=>ProxyのProxy Port",
  "ListenByClient": ":22",
  "comment": "ServerなのかProxyなのかを指定",
  "BootType": "Server"
}
```
## 動作
```mermaid
  sequenceDiagram
    autonumber
    actor Client
    Note over Client: Access<br>User
    participant Proxy
    participant ServerMachine
    participant EndPoint
    Note over EndPoint: ex)Minecraft Server<br>ARK Server
    ServerMachine->>Proxy: TCP request
    Client->>Proxy: TCP request
    loop I/O to I/O
      Client->>Proxy: Transfer Data
      Proxy->>ServerMachine: Transfer Data
      ServerMachine->>EndPoint: Transfer Data (TCP)
      EndPoint->>ServerMachine: Transfer Data (TCP)
      ServerMachine->>Proxy: Transfer Data
      Proxy->>Client: Transfer Data
    end
```

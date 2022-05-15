# go2tunnel
開発環境 : go version go1.17.8 linux/arm64 raspi4 8GB  

## how to use
1. go run main.go -env="???.json" で "BootServer":"Proxy" で起動  
2. go run main.go -env="???.json" で "BootServer":"Server" で起動  
この順番で起動すればPort開放がいらないはず?


## 変数 について
```json
{
  "comment":"使用するプロトコル",
  "UseProtcol":"tcp4", 
  "comment":" 実際のサーバーへアクセスするアドレス",
  "ServerLocalAddress":"localhost:25565",
  "comment":" プロキシと鯖を繋げる際に使うアドレス",
  "ProxyGlobalAddress":"example.com:80",
  "comment":" プロキシがListenするポート,ClientListenと異なること",
  "ProxyListen":":80",
  "comment":" クライアントがアクセスする際のポート,ProxyListenと異なること",
  "ClientListen":":22",
  "comment":"ServerなのかProxyなのかを指定",
  "BootServer": "Server"
}```
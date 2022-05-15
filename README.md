# go2tunnel
開発環境 : go version go1.17.8 linux/arm64 raspi4 8GB  

## how to use
1. main.goをProxy modeで起動
2. main.goをServer modeで起動
この順番で起動すればPort開放がいらないはず?

## 変数 について
```golang
var (
	UseProtcol                    = "tcp4"
	//使用するプロトコル
	ServerLocalAddress            = "localhost:25564"
	// 実際のサーバーへアクセスするアドレス
	ProxyGlobalAddress            = "example.com:25200"
	// プロキシと鯖を繋げる際に使うアドレス
	ProxyListen                   = ":25200"
	// プロキシがListenするポート
	ClientListen                  = ":25300"
	// クライアントがアクセスする際のポート
	BootServer         ServerType = Server
	// 以下の変数が指定可能
)
var (
	Server ServerType = 1
	Proxy  ServerType = 2
)```
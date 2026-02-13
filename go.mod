module aether-rea

go 1.26

require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/gorilla/websocket v1.5.3
	github.com/quic-go/quic-go v0.58.0
	github.com/quic-go/webtransport-go v0.8.0
	golang.org/x/crypto v0.35.0
	golang.org/x/time v0.12.0
)

replace github.com/quic-go/quic-go => ./third_party/quic-go

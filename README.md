# TradeGuildLedger

Use at your own risk, this project is in alpha

The website can be found at [TradeGuildLedger.com](https://www.TradeGuildLedger.com)

This is an addon for Elder Scrolls Online which simply takes Guild Trader listings and saves them locally on your machine. A client can then be run in order to upload this data to a server to display it.

The addon is written in Lua and the client and server are both written in Go. The project should run on Linux, MacOS or Windows. The client is currently only tested on Windows and the server is tested on Linux.

## Server

To build the server simply run
```bash
go build cmd/server/main.go
```

## Client

Build client for windows on mac
```bash
brew install mingw-w64
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui cmd/client/main.go
```

Build for linux on mac
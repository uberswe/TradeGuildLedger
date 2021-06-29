# TradeGuildLedger

Build for windows on mac
```bash
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build cmd/client/main.go
```

Build for linux on mac
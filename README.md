# TradeGuildLedger

*Archived since I have lost interest in ESO and moved on to other projects*

Use at your own risk, this project is in alpha.

The website can be found at [TradeGuildLedger.com](https://www.TradeGuildLedger.com)

This is an addon for Elder Scrolls Online which simply takes Guild Trader listings and saves them locally on your machine. A client can then be run in order to upload this data to a server to display it.

The addon is written in Lua and the client and server are both written in Go. The project should run on Linux, MacOS or Windows. The client is currently only tested on Windows and the server is tested on Linux.

Created by @uberswe (PC/EU), feel free to write me in-game if you have questions.

## Client

When running the client you can use the following env variables:
```
REMOTE_SERVER="https://www.tradeguildledger.com"
SAVED_VARIABLE_FILE="./path/to/saved/variables/file.lua"
```

### Build client for windows on mac
```bash
brew install mingw-w64
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui cmd/client/main.go
```

#### With CLI debug window
```bash
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build cmd/client/main.go
```

### Build client for windows on linux
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -o main.exe cmd/client/main.go
```

## Server

There is a docker-compose file which makes it easier to run the server. Simply run the following command:

```bash
docker-compose up
```

The server will now be built and once it is finished it will be available on port 8080.

### Build manually

When running the server you can use the following env variables:
```
HTTP_PORT=":3100"
```

The server is only tested on Ubuntu 18+. Modifications to code may be needed for other platforms. To build the server simply run
```bash
go build cmd/server/main.go
```

The html templates use [Bulma](https://bulma.io/), the easiest way to add this is with `npm install`

## Local development

I provided a docker-compose file to make it easy to run the server locally, use the following command to start the server:

```bash
docker-compose up --build -d
```

Your server will now run at [http://localhost:8080/](http://localhost:8080/) 

You can tail the logs with this command:
```bash
docker-compose logs -f
```

Then to run the client use the following command:

```bash
REMOTE_SERVER=http://localhost:8080 go run cmd/client/main.go 
```

## Addon

The addon is written in Lua and currently kept very basic. It reads listings when you view them in game and saves that data when the ui is reloaded or whenever the game triggers a write of saved variables. 

## Disclaimer

TradeGuildLedger is in no way related to Bethesda Softworks, ZeniMax Online Studios, or ZeniMax Media.

This Add-on is not created by, affiliated with or sponsored by ZeniMax Media Inc. or its affiliates. The Elder Scrolls® and related logos are registered trademarks or trademarks of ZeniMax Media Inc. in the United States and/or other countries. All rights reserved.

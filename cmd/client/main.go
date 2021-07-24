package main

import "github.com/uberswe/tradeguildledger/client"

var (
	Version   = "0.0.0"
	APIKey    = "DEV"
	BuildTime = ""
)

func main() {
	client.Run(Version, APIKey, BuildTime)
}

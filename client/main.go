package client

import (
	"fmt"
	"github.com/uberswe/tradeguildledger/pkg/parser"
	"log"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
)

var (
	logData   []string
	list      *widget.List
	url       = "http://localhost:3100"
	sv        = "savedvars/TradeGuildLedger.lua"
	version   = "0.0.0"
	apiKey    = "DEV"
	buildTime = ""
	lastLen   = 0
)

func Run(v string, a string, bt string) {
	version = v
	apiKey = a
	buildTime = bt
	log.Println("Running client version", version)
	log.Println("Build time", buildTime)
	log.Println("API key", apiKey)
	// TODO make compatible for other OSs
	if runtime.GOOS == "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		sv = fmt.Sprintf("%s\\Documents\\Elder Scrolls Online\\live\\SavedVariables\\TradeGuildLedger.lua", home)
		url = "https://www.tradeguildledger.com"
	}

	go parseLua()
	launchUI()
}

func parseLua() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)

					data, err := parser.LuaChunkParser(sv)
					if len(data.Listings) != lastLen {
						lastLen = len(data.Listings)
						if err != nil {
							log.Println(err)
						} else {
							go syncWithRemote(data)
						}
					} else {
						log.Println("number of listings is the same, skipping")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	failCount := 0

	for {
		err = watcher.Add(sv)
		if err != nil {
			log.Println(err)
			failCount++
		} else {
			data, err := parser.LuaChunkParser(sv)
			if err != nil {
				log.Println(err)
			} else {
				lastLen = len(data.Listings)
				go syncWithRemote(data)
			}
			break
		}
		time.Sleep(5 * time.Second)
	}

	<-done
}

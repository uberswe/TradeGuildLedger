package client

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	parser "github.com/uberswe/go-lua-table-parser"
)

var (
	logData   []string
	list      *widget.List
	url       = "http://localhost:3100"
	sv        = "savedvars/TradeGuildLedger.lua"
	version   = "0.0.0"
	apiKey    = "DEV"
	buildTime = ""
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

					mapResult, err := readFile(sv, 0)

					if err != nil {
						log.Println(err)
						break
					}

					go process(mapResult)
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
			break
		}
		time.Sleep(5 * time.Second)
	}

	<-done
}

func readFile(file string, attempt int) (map[string]interface{}, error) {
	log.Println("Attempting to read ", file)
	f, err := os.OpenFile(file, os.O_RDWR, os.FileMode(0666))
	if err != nil {
		if attempt < 10 {
			log.Println(err)
			time.Sleep(1 * time.Second)
			return readFile(file, attempt+1)
		}
		return nil, err
	}

	defer func() {
		if err = f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	b, err := parser.Parse(f, "TradeGuildLedgerVars")
	if err != nil {
		return nil, err
	}
	return b, nil
}

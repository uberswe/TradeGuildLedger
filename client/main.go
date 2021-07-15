package client

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	parser "github.com/uberswe/go-lua-table-parser"
)

var (
	logData  []string
	list     *widget.List
	checksum = ""
	url      = "http://localhost:3100"
	sv       = "savedvars/TradeGuildLedger.lua"
)

func Run() {
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

					s, err := readFile(sv, 0)

					if err != nil {
						log.Println(err)
						break
					}

					// Make a checksum of the ledger content so we only update changes
					hasher := sha1.New()
					hasher.Write(s)
					sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

					if sha == checksum {
						log.Println("no changes")
						break
					}
					addLog("Detected file write...")
					checksum = sha

					mapResult, err := parser.Parse(string(s), "TradeGuildLedgerVars")
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

func readFile(file string, attempt int) ([]byte, error) {
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

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

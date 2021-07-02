package client

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	json "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var (
	logData  []string
	list     *widget.List
	checksum = ""
)

func Run() {
	go parseLua()
	launchUI()
}

func parseLua() {
	url := "http://localhost:3100/api/v1/receive"
	sv := "savedvars/TradeGuildLedger.lua"
	if runtime.GOOS == "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			return
		}
		sv = fmt.Sprintf("%s\\Documents\\Elder Scrolls Online\\live\\SavedVariables\\TradeGuildLedger.lua", home)
		url = "https://www.tradeguildledger.com/api/v1/receive"
	}
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
					L := lua.NewState()
					defer L.Close()

					s, err := readFile(sv, 0)

					if err != nil {
						log.Println(err)
						return
					}

					// Make a checksum of the ledger content so we only update changes
					hasher := sha1.New()
					hasher.Write(s)
					sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

					if sha == checksum {
						return
					}
					checksum = sha
					if err := L.DoString(string(s)); err != nil {
						fmt.Println(err)
						return
					}
					lv := L.GetGlobal("TradeGuildLedgerVars")
					j, err := json.Encode(lv)
					if err != nil {
						fmt.Println(err)
						return
					}
					logData = append(logData, "Uploading data...")
					list.Refresh()

					fmt.Println("done parsing")

					fmt.Println("URL:>", url)

					req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					defer resp.Body.Close()

					fmt.Println("response Status:", resp.Status)
					fmt.Println("response Headers:", resp.Header)
					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Println("response Body:", string(body))
					logData = append(logData, fmt.Sprintf("Last uploaded at %s", time.Now().Format("2006-01-02 15:04:05")))
					list.Refresh()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(sv)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func readFile(file string, attempt int) ([]byte, error) {
	log.Println("Attempting to read ", file)
	f, err := os.Open(file)
	if err != nil {
		if attempt < 10 {
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

func launchUI() {
	logData = append(logData, "Application started, waiting to send data")
	a := app.NewWithID("com.tradeguildledger.app")
	w := a.NewWindow("Trade Guild Ledger Client")
	list = widget.NewList(
		func() int {
			return len(logData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(logData[i])
		})

	w.SetContent(list)
	w.Resize(fyne.NewSize(640, 460))

	w.ShowAndRun()
}

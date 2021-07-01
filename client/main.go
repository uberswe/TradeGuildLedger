package client

import (
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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
	status *widget.Label
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
		url = "https://tradeguildledger.com/api/v1/receive"
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

					if err := L.DoFile(sv); err != nil {
						fmt.Println(err)
						return
					}
					lv := L.GetGlobal("TradeGuildLedgerVars")
					j, err := json.Encode(lv)
					if err != nil {
						fmt.Println(err)
						return
					}
					status.SetText("Uploading data...")
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
					status.SetText(fmt.Sprintf("Last uploaded at %s", time.Now().Format("2006-01-02 15:04:05")))
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

func launchUI() {
	a := app.NewWithID("com.tradeguildledger.app")
	w := a.NewWindow("Trade Guild Ledger Client")
	status = widget.NewLabel("Waiting for changes.")
	w.SetContent(container.NewVBox(
		status,
	))
	w.Resize(fyne.NewSize(640, 460))

	w.ShowAndRun()
}

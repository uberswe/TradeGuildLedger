package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	json "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
	"log"
)

func main() {
	parseLua()
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
					L := lua.NewState()
					defer L.Close()
					if err := L.DoFile("savedvars/TradeGuildLedger.lua"); err != nil {
						fmt.Println(err)
						return
					}
					lv := L.GetGlobal("TradeGuildLedgerVars")
					_, err := json.Encode(lv)
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println("done parsing")
					//fmt.Println(string(j))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("savedvars/TradeGuildLedger.lua")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func launchUI() {
	a := app.NewWithID("com.tradeguildledger.app")
	w := a.NewWindow("Hello")
	newItem := fyne.NewMenuItem("New", nil)
	file := fyne.NewMenu("File", newItem)
	mainMenu := fyne.NewMainMenu(
		file,
	)
	w.SetMainMenu(mainMenu)
	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.Resize(fyne.NewSize(640, 460))

	w.ShowAndRun()
}

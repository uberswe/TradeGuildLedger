package client

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/emersion/go-autostart"
)

func launchUI() {
	// TODO move this to an about menu or something
	logData = append(logData, "TradeGuildLedger is not created by, affiliated with or sponsored by ZeniMax Media Inc. or its affiliates. The Elder Scrolls® and related logos are registered \ntrademarks or trademarks of ZeniMax Media Inc. in the United States and/or other countries. All rights reserved.")
	logData = append(logData, fmt.Sprintf("Watching file: %s", sv))
	logData = append(logData, fmt.Sprintf("Trade Guild Ledger Client %s\nhttps://www.tradeguildledger.com/\nForce update using /reloadui in game", version))

	a := app.NewWithID("com.tradeguildledger.app")
	w := a.NewWindow("Trade Guild Ledger Client")
	list = widget.NewList(
		func() int {
			return len(logData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template\nis\nmultiline\nnow")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).Wrapping = fyne.TextWrapWord
			o.(*widget.Label).SetText(logData[i])
		})
	checkbox := widget.NewCheck("Run Trade Guild Ledger Client at startup (buggy)", runAtStartup)
	w.SetContent(container.New(layout.NewBorderLayout(nil, checkbox, nil, nil), list, checkbox))
	w.Resize(fyne.NewSize(640, 460))

	w.ShowAndRun()
}

func runAtStartup(b bool) {
	ex, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(ex)

	// TODO make compatible for other OSs
	execstring := []string{"open", ex}

	if runtime.GOOS == "windows" {
		execstring = []string{"start", ex}
	}

	app := &autostart.App{
		Name:        "tradeguildledgerclient",
		DisplayName: "Trade Guild Ledger Client",
		Exec:        execstring,
	}

	if b && !app.IsEnabled() {
		if err := app.Enable(); err != nil {
			log.Println(err)
			return
		}
		log.Println("App run at startup enabled")
	} else if !b && app.IsEnabled() {
		if err := app.Disable(); err != nil {
			log.Println(err)
			return
		}
		log.Println("App run at startup disabled")
	}
}

func addLog(s string) {
	log.Println(s)
	logData = append(logData, s)
	if list != nil {
		list.Refresh()
		list.Select(len(logData) - 1)
	}
}

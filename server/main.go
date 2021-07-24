package server

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB
	addonVersion  = "0.0.1"
	serverVersion = "0.0.1"
	port          = ":3100"
)

type APIResponse struct {
	Message string
}

type IndexData struct {
	Updates       []UpdateModel
	ListingsCount int64
	UpdatesCount  int64
	ItemsCount    int64
}

type UpdateData struct {
	Updates    []UpdateModel
	Offset     int
	NextOffset int
	PrevOffset int
}

type NpcData struct {
	Npcs       []NpcModel
	Offset     int
	NextOffset int
	PrevOffset int
}

func Run() {
	// DB
	initDB()

	// Generate Downloads
	go buildAddonZip()
	go buildWindowsClient()

	// HTTP
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("./assets"))
	router.ServeFiles("/vendor/bulma/css/*filepath", http.Dir("./node_modules/bulma/css"))
	router.GET("/", index)
	router.GET("/downloads", downloads)
	router.GET("/downloads/:type", handleDownload)
	router.GET("/ledger/listings", listings)
	router.GET("/ledger/listings/:offset", listings)
	router.GET("/ledger/item/:slug", item)
	router.GET("/ledger/traders", traders)
	router.GET("/ledger/traders/:offset", traders)
	router.GET("/ledger/trader/:slug", trader)
	router.GET("/ledger/events", events)
	router.GET("/ledger/events/:offset", events)

	// API
	router.POST("/api/v1/receive", receive)

	router.POST("/api/v2/items", receiveItems)
	router.POST("/api/v2/listings", receiveListings)
	router.GET("/api/v2/items", fetchItems)
	router.GET("/api/v2/listings", fetchListings)
	router.GET("/api/v2/addon/version", fetchAddonVersion)

	log.Println(fmt.Sprintf("TradeGuildLedgerServer %s", serverVersion))
	log.Println(fmt.Sprintf("Listening on %s", port))
	log.Fatal(http.ListenAndServe(port, router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "index.html")
	var u []UpdateModel
	if res := db.Find(&u).Order("id asc").Limit(30); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var updatesCount int64
	var listingCount int64
	var itemsCount int64

	db.Model(&UpdateModel{}).Count(&updatesCount)
	db.Model(&ListingModel{}).Count(&listingCount)
	db.Model(&ItemModel{}).Count(&itemsCount)

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout", IndexData{
		Updates:       u,
		UpdatesCount:  updatesCount,
		ListingsCount: listingCount,
		ItemsCount:    itemsCount,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func downloads(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "downloads.html")

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	t := p.ByName("type")
	filePath := ""
	if t == "client" {
		w.Header().Set("Content-Disposition", "attachment; filename=TradeGuildLedgerClient.exe")
		w.Header().Set("Content-Type", "application/octet-stream")
		filePath, err = findFileWithExtension("./downloads", ".exe")
	} else if t == "addon" {
		w.Header().Set("Content-Disposition", "attachment; filename=TradeGuildLedger.zip")
		w.Header().Set("Content-Type", "application/zip")
		filePath, err = findFileWithExtension("./downloads", ".zip")
	}
	if err == nil && filePath != "" {
		f, err := os.Open(filePath) // For read access.
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, f)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

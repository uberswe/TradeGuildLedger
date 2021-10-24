package server

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	db            *gorm.DB
	addonVersion  = "0.0.2"
	serverVersion = "0.0.2"
	port          = ":3100"
)

type APIResponse struct {
	Message   string
	RequestID string
}

type BaseData struct {
	DarkModeLink func(string) string
	FormatLink   func(string, string) string
	DarkMode     func(string) bool
	Region       func(string) string
	URLPath      string
	Title        string
}

type PaginationData struct {
	Offset     int
	NextOffset int
	PrevOffset int
}

type IndexData struct {
	BaseData
	Updates       []UpdateModel
	ListingsCount int64
	UpdatesCount  int64
	ItemsCount    int64
}

type UpdateData struct {
	BaseData
	PaginationData
	Updates []UpdateModel
}

type NpcData struct {
	BaseData
	PaginationData
	Npcs []NpcModel
}

func env() {
	if envPort, isset := os.LookupEnv("HTTP_PORT"); isset {
		port = envPort
	}
}

func Run() {
	env()
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

	// EU region routes replicated
	router.GET("/eu/ledger/listings", listings)
	router.GET("/eu/ledger/listings/:offset", listings)
	router.GET("/eu/ledger/item/:slug", item)
	router.GET("/eu/ledger/traders", traders)
	router.GET("/eu/ledger/traders/:offset", traders)
	router.GET("/eu/ledger/trader/:slug", trader)
	router.GET("/eu/ledger/events", events)
	router.GET("/eu/ledger/events/:offset", events)

	// US region routes replicated
	router.GET("/us/ledger/listings", listings)
	router.GET("/us/ledger/listings/:offset", listings)
	router.GET("/us/ledger/item/:slug", item)
	router.GET("/us/ledger/traders", traders)
	router.GET("/us/ledger/traders/:offset", traders)
	router.GET("/us/ledger/trader/:slug", trader)
	router.GET("/us/ledger/events", events)
	router.GET("/us/ledger/events/:offset", events)

	// Darkmode replicated
	router.GET("/dark/", index)
	router.GET("/dark/downloads", downloads)
	router.GET("/dark/downloads/:type", handleDownload)
	router.GET("/dark/ledger/listings", listings)
	router.GET("/dark/ledger/listings/:offset", listings)
	router.GET("/dark/ledger/item/:slug", item)
	router.GET("/dark/ledger/traders", traders)
	router.GET("/dark/ledger/traders/:offset", traders)
	router.GET("/dark/ledger/trader/:slug", trader)
	router.GET("/dark/ledger/events", events)
	router.GET("/dark/ledger/events/:offset", events)

	// Darkmode EU region routes replicated
	router.GET("/dark/eu/ledger/listings", listings)
	router.GET("/dark/eu/ledger/listings/:offset", listings)
	router.GET("/dark/eu/ledger/item/:slug", item)
	router.GET("/dark/eu/ledger/traders", traders)
	router.GET("/dark/eu/ledger/traders/:offset", traders)
	router.GET("/dark/eu/ledger/trader/:slug", trader)
	router.GET("/dark/eu/ledger/events", events)
	router.GET("/dark/eu/ledger/events/:offset", events)

	// Darkmode US region routes replicated
	router.GET("/dark/us/ledger/listings", listings)
	router.GET("/dark/us/ledger/listings/:offset", listings)
	router.GET("/dark/us/ledger/item/:slug", item)
	router.GET("/dark/us/ledger/traders", traders)
	router.GET("/dark/us/ledger/traders/:offset", traders)
	router.GET("/dark/us/ledger/trader/:slug", trader)
	router.GET("/dark/us/ledger/events", events)
	router.GET("/dark/us/ledger/events/:offset", events)

	// API
	router.POST("/api/v1/receive", removed)

	router.POST("/api/v2/items", removed)
	router.POST("/api/v2/listings", removed)
	router.GET("/api/v2/items", removed)
	router.GET("/api/v2/listings", removed)
	router.GET("/api/v2/addon/version", fetchAddonVersion)

	router.POST("/api/v3/receive", receiveData)

	log.Println(fmt.Sprintf("TradeGuildLedgerServer %s", serverVersion))
	log.Println(fmt.Sprintf("Listening on %s", port))
	log.Fatal(http.ListenAndServe(port, router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "index.html")
	var u []UpdateModel
	if res := db.Order("id desc").Limit(30).Find(&u); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
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
	indexData := IndexData{
		Updates:       u,
		UpdatesCount:  updatesCount,
		ListingsCount: listingCount,
		ItemsCount:    itemsCount,
	}
	indexData.URLPath = r.URL.Path
	indexData.DarkMode = findDarkmode
	indexData.FormatLink = linkFormatter
	indexData.Region = findRegion
	indexData.DarkModeLink = darkModeLinkFormatter
	indexData.Title = "Home"

	err = tmpl.ExecuteTemplate(w, "layout", indexData)
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
	err = tmpl.ExecuteTemplate(w, "layout", BaseData{
		FormatLink:   linkFormatter,
		DarkMode:     findDarkmode,
		Region:       findRegion,
		DarkModeLink: darkModeLinkFormatter,
		URLPath:      r.URL.Path,
		Title:        "Downloads",
	})
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

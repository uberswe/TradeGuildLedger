package server

import (
	"errors"
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
	db *gorm.DB
)

type APIResponse struct {
	Message string
}

type IndexData struct {
	Updates       []UpdateModel
	ListingsCount int64
	UpdatesCount  int64
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

	// HTTP
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("./assets"))
	router.ServeFiles("/vendor/bulma/css/*filepath", http.Dir("./node_modules/bulma/css"))
	router.GET("/", index)
	router.GET("/downloads", downloads)
	router.GET("/downloads/:type", handleDownload)
	router.GET("/ledger/listings", listings)
	router.GET("/ledger/listings/:offset", listings)
	router.GET("/ledger/traders", traders)
	router.GET("/ledger/traders/:offset", traders)
	router.GET("/ledger/events", events)
	router.GET("/ledger/events/:offset", events)

	// API
	router.POST("/api/v1/receive", receive)

	router.POST("/api/v2/items", receiveItems)
	router.POST("/api/v2/listings", receiveListings)
	router.GET("/api/v2/items", fetchItems)
	router.GET("/api/v2/listings", fetchListings)

	log.Println("Listening on :3100")
	log.Fatal(http.ListenAndServe(":3100", router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "index.html")
	var u []UpdateModel
	if res := db.Find(&u).Order("id desc").Limit(30); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var updatesCount int64
	var listingCount int64

	db.Model(&UpdateModel{}).Count(&updatesCount)
	db.Model(&ListingModel{}).Count(&listingCount)

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout", IndexData{
		Updates:       u,
		UpdatesCount:  updatesCount,
		ListingsCount: listingCount,
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
	return
}

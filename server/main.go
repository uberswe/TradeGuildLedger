package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type ListingData struct {
	Listings   []ListingModel
	Offset     int
	NextOffset int
	PrevOffset int
	Search     string
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

func listings(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "listings.html")

	queryValues := r.URL.Query()
	search := queryValues.Get("search")

	limit := 100

	offsetCount := 0
	offset := p.ByName("offset")
	if offset != "" {
		i, err := strconv.Atoi(offset)
		if err != nil {
			// handle error
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		offsetCount = i
	}

	// TODO add pagination
	var listings []ListingModel
	if res := db.Preload("ItemModel").
		Preload("NpcModel").
		Preload("SellerModel").
		Joins("left join item_models on listing_models.item_model_id = item_models.id").
		Where("item_models.name LIKE ?", fmt.Sprintf("%%%s%%", search)).
		Order("listing_models.id desc").
		Offset(offsetCount * limit).
		Limit(limit).
		Find(&listings); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", ListingData{
		Listings:   listings,
		Offset:     offsetCount,
		NextOffset: offsetCount + 1,
		PrevOffset: offsetCount - 1,
		Search:     search,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func events(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "updates.html")

	limit := 20

	offsetCount := 0
	offset := p.ByName("offset")
	if offset != "" {
		i, err := strconv.Atoi(offset)
		if err != nil {
			// handle error
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		offsetCount = i
	}

	var updates []UpdateModel
	if res := db.Offset(offsetCount * limit).
		Limit(limit).
		Order("id desc").
		Find(&updates); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout", UpdateData{
		Updates:    updates,
		Offset:     offsetCount,
		NextOffset: offsetCount + 1,
		PrevOffset: offsetCount - 1,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func traders(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "traders.html")

	limit := 20

	offsetCount := 0
	offset := p.ByName("offset")
	if offset != "" {
		i, err := strconv.Atoi(offset)
		if err != nil {
			// handle error
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		offsetCount = i
	}

	var npcs []NpcModel
	if res := db.Offset(offsetCount * limit).
		Limit(limit).
		Order("id desc").
		Find(&npcs); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout", NpcData{
		Npcs:       npcs,
		Offset:     offsetCount,
		NextOffset: offsetCount + 1,
		PrevOffset: offsetCount - 1,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func receive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("received data")
	p, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	go parseIncomingPayload(p, r)
	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(APIResponse{
		Message: "received",
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
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

func findFileWithExtension(folder string, extension string) (string, error) {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, extension) {
			return file, nil
		}
	}
	return "", errors.New("no file could be found")
}

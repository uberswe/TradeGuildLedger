package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

var (
	db *gorm.DB
)

type APIResponse struct {
	Message string
}

func Run() {
	// DB
	initDB()

	// HTTP
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("./assets"))
	router.ServeFiles("/vendor/bulma/css/*filepath", http.Dir("./node_modules/bulma/css"))
	router.GET("/", index)
	router.GET("/downloads", index)
	router.GET("/ledger/listings", index)
	router.GET("/ledger/traders", index)
	router.GET("/ledger/events", index)

	// API
	router.POST("/api/v1/receive", receive)

	log.Println("Listening on :3100")
	log.Fatal(http.ListenAndServe(":3100", router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "index.html")

	tmpl, _ := template.ParseFiles(lp)
	_ = tmpl.ExecuteTemplate(w, "layout", nil)
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

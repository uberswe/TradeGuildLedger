package server

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gorm.io/driver/sqlite"
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

func Run() {
	var err error
	// DB
	db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// HTTP
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("./assets"))
	router.ServeFiles("/vendor/bulma/css/*filepath", http.Dir("./node_modules/bulma/css"))
	router.GET("/", index)
	router.POST("/api/v1/receive", receive)

	log.Println("Listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("web", "index.html")

	tmpl, _ := template.ParseFiles(lp)
	_ = tmpl.ExecuteTemplate(w, "layout", nil)
}

func receive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("received data")
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println("request Body:", string(body))
}

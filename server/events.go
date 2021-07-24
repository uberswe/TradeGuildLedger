package server

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

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
		Order("id asc").
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

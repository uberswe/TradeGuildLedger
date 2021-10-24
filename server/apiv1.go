package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func removed(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusInternalServerError)
}

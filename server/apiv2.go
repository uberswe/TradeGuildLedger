package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func fetchAddonVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	jData, err := json.Marshal(APIResponse{
		Message: addonVersion,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}

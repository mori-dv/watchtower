package api

import (
	"log"
	"net/http"
)
func HealthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Println("health was not write on response body successfully")
	}
}

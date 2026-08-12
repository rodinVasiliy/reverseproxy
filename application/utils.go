package application

import (
	"fmt"
	"log"
	"net/http"
)

func deny(w http.ResponseWriter) {
	http.Error(w, "Access denied", http.StatusForbidden)
}

func notFound(w http.ResponseWriter) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

func internalError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func fail(msg string, err error) {
	fmt.Printf("%s: %v", msg, err)
	log.Fatal(err)
}

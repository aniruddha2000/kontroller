package handlers

import (
	"fmt"
	"net/http"
)

func KlusterValidationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("KlusterValidationHandler is called")
}

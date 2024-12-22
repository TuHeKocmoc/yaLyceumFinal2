package main

import (
	"net/http"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
)

func main() {
	// запускаем сервер
	http.HandleFunc("/api/v1/calculate", handler.CalcHandler)
	http.ListenAndServe(":8080", nil)
}

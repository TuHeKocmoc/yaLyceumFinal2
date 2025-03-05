package main

import (
	"log"
	"net/http"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
)

func main() {
	http.HandleFunc("/api/v1/calculate", handler.HandleCreateExpression)    // POST
	http.HandleFunc("/api/v1/expressions", handler.HandleGetAllExpressions) // GET
	http.HandleFunc("/api/v1/expressions/", handler.HandleGetExpressionByID)

	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.HandleGetTask(w, r)
		} else if r.Method == http.MethodPost {
			handler.HandlePostTaskResult(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Starting server at :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

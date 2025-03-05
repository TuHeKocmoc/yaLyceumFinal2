package main

import (
	"log"
	"net/http"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
)

func main() {
	if err := handler.InitTemplates(); err != nil {
		log.Fatalf("cannot init templates: %v", err)
	}

	http.HandleFunc("/", handler.HandleFrontIndex)        // GET
	http.HandleFunc("/front/add", handler.HandleFrontAdd) // POST
	http.HandleFunc("/expression/", handler.HandleFrontExpression)

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

	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Starting server at :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

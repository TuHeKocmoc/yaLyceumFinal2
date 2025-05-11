package main

import (
	"log"
	"net/http"
	"os"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/grpcserver"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
)

func main() {

	if err := db.InitDB(); err != nil {
		log.Fatalf("cannot init DB: %v", err)
	}

	if err := handler.InitTemplates(); err != nil {
		log.Fatalf("cannot init templates: %v", err)
	}

	go func() {
		grpcAddr := os.Getenv("GRPC_ADDR")
		if grpcAddr == "" {
			grpcAddr = ":50051"
		}
		grpcserver.StartGRPCServer(grpcAddr)
	}()

	http.HandleFunc("/api/v1/register", handler.HandleRegister)
	http.HandleFunc("/api/v1/login", handler.HandleLogin)

	http.HandleFunc("/", handler.HandleFrontIndex)
	http.HandleFunc("/front/add", handler.HandleFrontAdd)
	http.HandleFunc("/expression/", handler.HandleFrontExpression)

	// 6) Защищённые эндпоинты — AuthMiddleware
	http.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)))
	http.Handle("/api/v1/expressions",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetAllExpressions)))
	http.Handle("/api/v1/expressions/",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetExpressionByID)))

	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("[MAIN] Starting HTTP server at :" + port + "...")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"api-server/internal/api"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting API server...")
	router := api.SetupRoutes()
	log.Fatal(http.ListenAndServe(":8080", router))
}

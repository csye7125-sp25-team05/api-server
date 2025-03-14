package main

import (
	"api-server/internal/api"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Starting API server...")
	db, err := sql.Open("postgres", "user=postgres dbname=api options=-csearch_path=api,public sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	router := api.SetupRoutes(db)
	log.Fatal(http.ListenAndServe(":8080", router))
}

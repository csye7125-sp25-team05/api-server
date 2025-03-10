package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers).Methods("GET")
	return router
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Send the user as a JSON response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
	}
}

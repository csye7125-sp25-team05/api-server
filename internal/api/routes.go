package api

import (
	"database/sql"
	"log"
	"net/http"

	"api-server/internal/handlers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func BasicAuth(db *sql.DB) func(http.Handler) http.Handler {
	// Ensure the user table exists in the api schema
	err := ensureUsersTable(db)
	if err != nil {
		log.Printf("Warning: Could not ensure api.user table exists: %v", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var storedPassword string
			// Query the api.user table instead of users
			err := db.QueryRow("SELECT password FROM api.user WHERE username = $1", username).Scan(&storedPassword)
			switch {
			case err == sql.ErrNoRows:
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			case err != nil:
				log.Printf("Database error during auth: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if password != storedPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ensureUsersTable creates the user table in the api schema if it doesn't exist
func ensureUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS api.user (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			account_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			account_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Optional: Insert a default admin user for testing
	_, err = db.Exec(`
		INSERT INTO api.user (username, first_name, last_name, password)
		SELECT 'admin@example.com', 'Admin', 'User', 'admin123'
		WHERE NOT EXISTS (SELECT 1 FROM api.user WHERE username = 'admin@example.com')
	`)
	return err
}

func SetupRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint (no BasicAuth)
	router.HandleFunc("/health", HealthCheckHandler(db)).Methods("GET")

	// User POST endpoint for creating new users (no BasicAuth)
	userHandler := handlers.NewUserHandler(db)
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

	// Create a subrouter for protected routes with BasicAuth
	authRouter := mux.NewRouter()
	authRouter.Use(BasicAuth(db))

	// Instructor Routes
	ir := handlers.NewInstructorHandler(db)
	authRouter.HandleFunc("/instructors", ir.GetInstructors).Methods("GET")
	authRouter.HandleFunc("/instructors/{id}", ir.GetInstructorByID).Methods("GET")
	authRouter.HandleFunc("/instructors", ir.CreateInstructor).Methods("POST")
	authRouter.HandleFunc("/instructors/{id}", ir.UpdateInstructor).Methods("PUT")
	authRouter.HandleFunc("/instructors/{id}", ir.DeleteInstructor).Methods("DELETE")

	// Course Routes
	courseHandler := handlers.NewCourseHandler(db)
	authRouter.HandleFunc("/courses", courseHandler.GetCourses).Methods("GET")
	authRouter.HandleFunc("/courses/{id}", courseHandler.GetCourseByID).Methods("GET")
	authRouter.HandleFunc("/courses", courseHandler.CreateCourse).Methods("POST")
	authRouter.HandleFunc("/courses/{id}", courseHandler.UpdateCourse).Methods("PUT")
	authRouter.HandleFunc("/courses/{id}", courseHandler.DeleteCourse).Methods("DELETE")

	// User Routes (excluding POST which is defined above without auth)
	authRouter.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	authRouter.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")
	authRouter.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	authRouter.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Trace Routes
	traceHandler := handlers.NewTraceHandler(db)
	authRouter.HandleFunc("/traces", traceHandler.GetTraces).Methods("GET")
	authRouter.HandleFunc("/traces/{id}", traceHandler.GetTraceByID).Methods("GET")
	authRouter.HandleFunc("/traces", traceHandler.CreateTrace).Methods("POST")
	authRouter.HandleFunc("/traces/{id}", traceHandler.UpdateTrace).Methods("PUT")
	authRouter.HandleFunc("/traces/{id}", traceHandler.DeleteTrace).Methods("DELETE")

	// Mount the authRouter under the main router
	router.PathPrefix("/").Handler(authRouter)

	return router
}

// HealthCheckHandler returns the health status of the application
func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

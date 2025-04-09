package api

import (
	"database/sql"
	"net/http"
	"os"

	"api-server/internal/handlers"

	"github.com/gorilla/mux"
)

func BasicAuth(next http.Handler) http.Handler {
	username := os.Getenv("BASIC_AUTH_USERNAME")
	password := os.Getenv("BASIC_AUTH_PASSWORD")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetupRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint (no BasicAuth)
	router.HandleFunc("/health", HealthCheckHandler(db)).Methods("GET")

	// Create a subrouter for instructor, course, user, and trace routes that require BasicAuth
	authRouter := router.PathPrefix("/instructors").Subrouter()

	// Apply BasicAuth middleware only to GET, PUT, DELETE methods for the subrouter
	authRouter.Use(BasicAuth) // This will apply BasicAuth to all routes under /instructors, including GET, PUT, DELETE

	// Instructor Routes (BasicAuth is already applied to authRouter)
	ir := handlers.NewInstructorHandler(db)
	authRouter.HandleFunc("", ir.GetInstructors).Methods("GET")
	authRouter.HandleFunc("/{id}", ir.GetInstructorByID).Methods("GET")
	authRouter.HandleFunc("", ir.CreateInstructor).Methods("POST") // POST does not need BasicAuth
	authRouter.HandleFunc("/{id}", ir.UpdateInstructor).Methods("PUT")
	authRouter.HandleFunc("/{id}", ir.DeleteInstructor).Methods("DELETE")

	// Course Routes (apply BasicAuth to GET, PUT, DELETE methods)
	courseHandler := handlers.NewCourseHandler(db)
	authRouter.HandleFunc("/courses", courseHandler.GetCourses).Methods("GET")
	authRouter.HandleFunc("/courses/{id}", courseHandler.GetCourseByID).Methods("GET")
	authRouter.HandleFunc("/courses", courseHandler.CreateCourse).Methods("POST") // POST does not need BasicAuth
	authRouter.HandleFunc("/courses/{id}", courseHandler.UpdateCourse).Methods("PUT")
	authRouter.HandleFunc("/courses/{id}", courseHandler.DeleteCourse).Methods("DELETE")

	// User Routes (apply BasicAuth to GET, PUT, DELETE methods)
	userHandler := handlers.NewUserHandler(db)
	authRouter.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	authRouter.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")
	authRouter.HandleFunc("/users", userHandler.CreateUser).Methods("POST") // POST does not need BasicAuth
	authRouter.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	authRouter.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Trace Routes (apply BasicAuth to GET, PUT, DELETE methods)
	traceHandler := handlers.NewTraceHandler(db)
	authRouter.HandleFunc("/traces", traceHandler.GetTraces).Methods("GET")
	authRouter.HandleFunc("/traces/{id}", traceHandler.GetTraceByID).Methods("GET")
	authRouter.HandleFunc("/traces", traceHandler.CreateTrace).Methods("POST") // POST does not need BasicAuth
	authRouter.HandleFunc("/traces/{id}", traceHandler.UpdateTrace).Methods("PUT")
	authRouter.HandleFunc("/traces/{id}", traceHandler.DeleteTrace).Methods("DELETE")

	return router
}

// HealthCheckHandler returns the health status of the application
func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := db.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		// Optionally, add more health checks (e.g., GCS connectivity)
		// For now, just return 200 OK if the database is reachable
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

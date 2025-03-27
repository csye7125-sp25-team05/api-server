package api

import (
	"database/sql"
	"net/http"

	"api-server/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", HealthCheckHandler(db)).Methods("GET")

	// Instructor Routes
	ir := handlers.NewInstructorHandler(db)
	router.HandleFunc("/instructors", ir.GetInstructors).Methods("GET")
	router.HandleFunc("/instructors/{id}", ir.GetInstructorByID).Methods("GET")
	router.HandleFunc("/instructors", ir.CreateInstructor).Methods("POST")
	router.HandleFunc("/instructors/{id}", ir.UpdateInstructor).Methods("PUT")
	router.HandleFunc("/instructors/{id}", ir.DeleteInstructor).Methods("DELETE")

	// Course Routes
	courseHandler := handlers.NewCourseHandler(db)
	router.HandleFunc("/courses", courseHandler.GetCourses).Methods("GET")
	router.HandleFunc("/courses/{id}", courseHandler.GetCourseByID).Methods("GET")
	router.HandleFunc("/courses", courseHandler.CreateCourse).Methods("POST")
	router.HandleFunc("/courses/{id}", courseHandler.UpdateCourse).Methods("PUT")
	router.HandleFunc("/courses/{id}", courseHandler.DeleteCourse).Methods("DELETE")

	// User Routes
	userHandler := handlers.NewUserHandler(db)
	router.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Trace Routes
	traceHandler := handlers.NewTraceHandler(db)
	router.HandleFunc("/traces", traceHandler.GetTraces).Methods("GET")
	router.HandleFunc("/traces/{id}", traceHandler.GetTraceByID).Methods("GET")
	router.HandleFunc("/traces", traceHandler.CreateTrace).Methods("POST")
	router.HandleFunc("/traces/{id}", traceHandler.UpdateTrace).Methods("PUT")
	router.HandleFunc("/traces/{id}", traceHandler.DeleteTrace).Methods("DELETE")

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

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"api-server/internal/model"
	"api-server/internal/repository"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type TraceHandler struct {
	tr                  *repository.TraceRepository
	ctx                 context.Context
	client              *storage.Client
	config              *Config
	serviceAccountEmail string
}

type Config struct {
	BucketName            string
	ServiceAccountKeyPath string
	Environment           string // "local" or "gke"
	ProjectID             string
}

func LoadConfig() *Config {
	// Try to load from .env file, but don't error if not found
	_ = godotenv.Load()

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local" // Default to local if not specified
	}

	return &Config{
		BucketName:            os.Getenv("BUCKET_NAME"),
		ServiceAccountKeyPath: os.Getenv("SERVICE_ACCOUNT_KEY_PATH"),
		Environment:           environment,
		ProjectID:             os.Getenv("PROJECT_ID"),
	}
}

func NewTraceHandler(db *sql.DB) *TraceHandler {
	ctx := context.Background()
	config := LoadConfig()

	var client *storage.Client
	var err error

	// Log environment for debugging
	fmt.Printf("Environment: %s, Bucket: %s\n", config.Environment, config.BucketName)

	if config.Environment == "local" {
		if config.ServiceAccountKeyPath != "" {
			client, err = storage.NewClient(ctx, option.WithCredentialsFile(config.ServiceAccountKeyPath))
		} else {
			client, err = storage.NewClient(ctx) // Fallback to ADC
		}
	} else {
		// GKE: Rely on Workload Identity (default credentials from metadata server)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to create storage client: %v", err))
	}

	return &TraceHandler{
		tr:     repository.NewTraceRepository(db),
		ctx:    ctx,
		client: client,
		config: config,
	}
}

// func NewTraceHandler(db *sql.DB) *TraceHandler {
// 	ctx := context.Background()
// 	config := LoadConfig()

// 	var client *storage.Client
// 	var err error
// 	var serviceAccountEmail string

// 	if config.Environment == "local" {
// 		// For local development, use service account key file
// 		if config.ServiceAccountKeyPath != "" {
// 			client, err = storage.NewClient(ctx, option.WithCredentialsFile(config.ServiceAccountKeyPath))
// 		} else {
// 			// Fallback to application default credentials
// 			client, err = storage.NewClient(ctx)
// 		}
// 	} else {
// 		// For GKE environment, use Workload Identity (no explicit credentials needed)
// 		client, err = storage.NewClient(ctx)
// 		serviceAccountEmail = os.Getenv("K8S_SERVICE_ACCOUNT_EMAIL")
// 	}

// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to create storage client: %v", err))
// 	}

// 	return &TraceHandler{tr: repository.NewTraceRepository(db), ctx: ctx, client: client, config: config,
// 		serviceAccountEmail: serviceAccountEmail,
// 	}
// }

func (th *TraceHandler) GetTraces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	traces, err := th.tr.GetAllTraces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(traces)
}

func (th *TraceHandler) GetTraceByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	trace, err := th.tr.GetTraceByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(trace)
}

func (th *TraceHandler) CreateTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Filename == "" || !strings.HasSuffix(header.Filename, ".pdf") {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	userIDStr := r.FormValue("user_id")
	var userID uuid.UUID
	if userIDStr == "" {
		userID = uuid.New()
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id format", http.StatusBadRequest)
			return
		}
	}

	bucketName := th.config.BucketName
	objectName := uuid.New().String() + filepath.Ext(header.Filename)
	fmt.Printf("Uploading to bucket: %s, object: %s\n", bucketName, objectName)

	wc := th.client.Bucket(bucketName).Object(objectName).NewWriter(th.ctx)
	wc.ContentType = header.Header.Get("Content-Type")
	if _, err = io.Copy(wc, file); err != nil {
		fmt.Printf("Copy error: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}
	if err := wc.Close(); err != nil {
		fmt.Printf("Close error: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to finalize upload: %v", err), http.StatusInternalServerError)
		return
	}

	attrs, err := th.client.Bucket(bucketName).Object(objectName).Attrs(th.ctx)
	if err != nil {
		fmt.Printf("Attrs error: %v\n", err)
		fmt.Printf("Attrs: %v\n", attrs)
		http.Error(w, fmt.Sprintf("Failed to get object attributes: %v", err), http.StatusInternalServerError)
		return
	}
	bucketPath := fmt.Sprintf("gs://%s/%s", bucketName, objectName)

	var trace model.Trace
	trace.UserID = userID
	trace.FileName = header.Filename
	trace.BucketPath = bucketPath
	err = th.tr.CreateTrace(&trace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trace)
}

// func (th *TraceHandler) CreateTrace(w http.ResponseWriter, r *http.Request) {
// 	// Handle file upload
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// Check if the file is a PDF
// 	if header.Filename == "" || !strings.HasSuffix(header.Filename, ".pdf") {
// 		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
// 		return
// 	}

// 	// Get user ID from form or generate a new one if empty
// 	userIDStr := r.FormValue("user_id")
// 	var userID uuid.UUID
// 	if userIDStr == "" {
// 		// Generate a new UUID if not provided
// 		userID = uuid.New()
// 	} else {
// 		// Try to parse the provided UUID
// 		var err error
// 		userID, err = uuid.Parse(userIDStr)
// 		if err != nil {
// 			http.Error(w, "Invalid user_id format", http.StatusBadRequest)
// 			return
// 		}
// 	}

// 	// Upload file to GCS
// 	bucketName := th.config.BucketName
// 	objectName := uuid.New().String() + filepath.Ext(header.Filename)
// 	wc := th.client.Bucket(bucketName).Object(objectName).NewWriter(th.ctx)
// 	wc.ContentType = header.Header.Get("Content-Type")
// 	if _, err = io.Copy(wc, file); err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	if err := wc.Close(); err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to finalize upload: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Get the URL of the uploaded file
// 	attrs, err := th.client.Bucket(bucketName).Object(objectName).Attrs(th.ctx)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to get object attributes: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	fmt.Printf("Object Attributes: %+v\n", attrs)
// 	bucketPath := fmt.Sprintf("gs://%s/%s", bucketName, objectName)

// 	// Create trace entry
// 	var trace model.Trace
// 	trace.UserID = userID
// 	trace.FileName = header.Filename
// 	trace.BucketPath = bucketPath
// 	err = th.tr.CreateTrace(&trace)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(trace)
// }

func (th *TraceHandler) UpdateTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var trace model.Trace
	err = json.NewDecoder(r.Body).Decode(&trace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = th.tr.UpdateTrace(id, &trace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (th *TraceHandler) DeleteTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// First get the trace to find the file location
	trace, err := th.tr.GetTraceByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Delete the file from GCS if it exists
	if trace.BucketPath != "" {
		// Parse bucket path to get object name
		parts := strings.Split(trace.BucketPath, "/")
		if len(parts) > 2 {
			objectName := parts[len(parts)-1]
			bucketName := th.config.BucketName

			// Delete the object
			err = th.client.Bucket(bucketName).Object(objectName).Delete(th.ctx)
			if err != nil {
				fmt.Printf("Warning: Failed to delete object from storage: %v\n", err)
				// Continue with trace deletion even if file deletion failed
			}
		}
	}

	// Delete from database
	err = th.tr.DeleteTrace(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

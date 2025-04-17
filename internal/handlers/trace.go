package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"api-server/internal/model"
	"api-server/internal/repository"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/api/option"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
)

type TraceHandler struct {
	tr     *repository.TraceRepository
	ctx    context.Context
	client *storage.Client
	config *Config
	logger *logging.Logger
}

type Config struct {
	BucketName            string
	ServiceAccountKeyPath string
	Environment           string // "local" or "gke"
	ProjectID             string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
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
	var logger *logging.Logger
	var err error

	// Debug: Log config loading
	log.Printf("DEBUG: Loading configuration - Environment: %s, ProjectID: %s", config.Environment, config.ProjectID)

	logClient, err := logging.NewClient(ctx, config.ProjectID)
	if err != nil {
		log.Printf("ERROR: Failed to create logging client, falling back to stdout: %v", err)
	} else {
		logger = logClient.Logger("trace-handler-logs", logging.CommonResource(&monitoredrespb.MonitoredResource{
			Type: "k8s_container",
			Labels: map[string]string{
				"project_id":     config.ProjectID,
				"namespace_name": "api-server",
				"container_name": "api-server",
				"cluster_name":   "my-gke-cluster",
				"location":       "us-east1",
				"pod_name":       os.Getenv("HOSTNAME"),
			},
		}))
		// Debug: Log successful logger initialization
		logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  "Logger initialized successfully for trace-handler-logs",
		})
	}

	// Info: Log environment details
	if logger != nil {
		logger.Log(logging.Entry{
			Severity: logging.Info,
			Payload:  "Initializing TraceHandler - Environment: " + config.Environment + ", Bucket: " + config.BucketName,
		})
	} else {
		log.Printf("INFO: Initializing TraceHandler - Environment: %s, Bucket: %s", config.Environment, config.BucketName)
	}

	// Storage client setup
	if config.Environment == "local" {
		logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  "Setting up storage client for local environment",
		})
		if config.ServiceAccountKeyPath != "" {
			client, err = storage.NewClient(ctx, option.WithCredentialsFile(config.ServiceAccountKeyPath))
		} else {
			client, err = storage.NewClient(ctx)
		}
	} else {
		logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  "Setting up storage client for GKE with Workload Identity",
		})
		client, err = storage.NewClient(ctx)
	}
	if err != nil {
		if logger != nil {
			logger.Log(logging.Entry{
				Severity: logging.Critical,
				Payload:  fmt.Sprintf("Failed to create storage client: %v", err),
			})
		}
		log.Fatal("CRITICAL: Failed to create storage client:", err)
	} else {
		logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  "Storage client initialized successfully",
		})
	}

	return &TraceHandler{
		tr:     repository.NewTraceRepository(db),
		ctx:    ctx,
		client: client,
		config: config,
		logger: logger,
	}
}

func (th *TraceHandler) GetTraces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  "Fetching all traces",
	})

	// Debug: Log request details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Received GetTraces request from %s", r.RemoteAddr),
	})

	traces, err := th.tr.GetAllTraces()
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to fetch traces: %v", err),
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Successfully fetched %d traces", len(traces)),
	})
	// Debug: Log response preparation
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  "Encoding traces to JSON response",
	})
	json.NewEncoder(w).Encode(traces)
}

func (th *TraceHandler) GetTraceByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Fetching trace with ID: %s", idStr),
	})
	// Debug: Log request details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Received GetTraceByID request for ID %s from %s", idStr, r.RemoteAddr),
	})

	id, err := uuid.Parse(idStr)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Invalid trace ID format: %v", err),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trace, err := th.tr.GetTraceByID(id)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to fetch trace with ID %s: %v", idStr, err),
		})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Successfully fetched trace with ID: %s", idStr),
	})
	// Debug: Log response preparation
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Encoding trace %s to JSON response", idStr),
	})
	json.NewEncoder(w).Encode(trace)
}

func (th *TraceHandler) CreateTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  "Starting trace creation",
	})

	// Start a new span
	ctx, span := otel.Tracer("api-server").Start(r.Context(), "CreateTrace")
	defer span.End()

	// Debug: Log request details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Received CreateTrace request from %s", r.RemoteAddr),
	})

	file, header, err := r.FormFile("file")
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Failed to get file from form: %v", err),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Filename == "" || !strings.HasSuffix(header.Filename, ".pdf") {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Invalid file: must be a PDF, got: %s", header.Filename),
		})
		span.SetStatus(codes.Error, "Invalid file: must be a PDF")
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}
	// Add file attributes to the span
	span.SetAttributes(
		attribute.String("file.name", header.Filename),
		attribute.Int64("file.size", header.Size),
	)

	// Debug: Log file details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Received file: %s, size: %d", header.Filename, header.Size),
	})

	userIDStr := r.FormValue("user_id")
	var userID uuid.UUID
	if userIDStr == "" {
		userID = uuid.New()
		th.logger.Log(logging.Entry{
			Severity: logging.Info,
			Payload:  fmt.Sprintf("Generated new user ID: %s", userID.String()),
		})
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			th.logger.Log(logging.Entry{
				Severity: logging.Warning,
				Payload:  fmt.Sprintf("Invalid user_id format: %v", err),
			})
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			http.Error(w, "Invalid user_id format", http.StatusBadRequest)
			return
		}
		// Debug: Log parsed user ID
		th.logger.Log(logging.Entry{
			Severity: logging.Debug,
			Payload:  fmt.Sprintf("Parsed user ID: %s", userID.String()),
		})
	}

	bucketName := th.config.BucketName
	objectName := uuid.New().String() + filepath.Ext(header.Filename)
	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Uploading to bucket: %s, object: %s", bucketName, objectName),
	})

	// Create a child span for GCS upload
	_, uploadSpan := otel.Tracer("api-server").Start(ctx, "UploadToGCS")
	wc := th.client.Bucket(bucketName).Object(objectName).NewWriter(th.ctx)
	wc.ContentType = header.Header.Get("Content-Type")
	// Debug: Log upload start
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Starting upload to GCS: %s/%s", bucketName, objectName),
	})
	if _, err = io.Copy(wc, file); err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to upload file to GCS: %v", err),
		})
		uploadSpan.RecordError(err)
		uploadSpan.SetStatus(codes.Error, err.Error())
		uploadSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}
	if err := wc.Close(); err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to finalize GCS upload: %v", err),
		})
		uploadSpan.RecordError(err)
		uploadSpan.SetStatus(codes.Error, err.Error())
		uploadSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, fmt.Sprintf("Failed to finalize upload: %v", err), http.StatusInternalServerError)
		return
	}
	uploadSpan.End()

	// Debug: Log upload completion
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Completed upload to GCS: %s/%s", bucketName, objectName),
	})

	attrs, err := th.client.Bucket(bucketName).Object(objectName).Attrs(th.ctx)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to get GCS object attributes: %v", err),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, fmt.Sprintf("Failed to get object attributes: %v", err), http.StatusInternalServerError)
		return
	}
	// Debug: Log object attributes
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Fetched GCS object attributes: size=%d, content_type=%s", attrs.Size, attrs.ContentType),
	})
	bucketPath := fmt.Sprintf("gs://%s/%s", bucketName, objectName)

	var trace model.Trace
	trace.UserID = userID
	trace.FileName = header.Filename
	trace.BucketPath = bucketPath

	// Debug: Log trace creation attempt
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Creating trace in database: user_id=%s, filename=%s", trace.UserID, trace.FileName),
	})

	// Create a child span for database operation
	_, dbSpan := otel.Tracer("api-server").Start(ctx, "CreateTraceDB")
	err = th.tr.CreateTrace(&trace)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to create trace in database: %v", err),
		})
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dbSpan.End()

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Successfully created trace with ID: %s", trace.ID.String()),
	})
	// Debug: Log response preparation
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Encoding trace %s to JSON response", trace.ID.String()),
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trace)
}

func (th *TraceHandler) UpdateTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Updating trace with ID: %s", idStr),
	})
	// Debug: Log request details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("_received UpdateTrace request for ID %s from %s", idStr, r.RemoteAddr),
	})

	id, err := uuid.Parse(idStr)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Invalid trace ID format: %v", err),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var trace model.Trace
	// Debug: Log decoding attempt
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Decoding update request body for trace %s", idStr),
	})
	err = json.NewDecoder(r.Body).Decode(&trace)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Failed to decode trace update request: %v", err),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = th.tr.UpdateTrace(id, &trace)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to update trace with ID %s: %v", idStr, err),
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Successfully updated trace with ID: %s", idStr),
	})
	// Debug: Log response completion
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Completed update for trace %s, returning no content", idStr),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (th *TraceHandler) DeleteTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Deleting trace with ID: %s", idStr),
	})
	// Debug: Log request details
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Received DeleteTrace request for ID %s from %s", idStr, r.RemoteAddr),
	})

	id, err := uuid.Parse(idStr)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Warning,
			Payload:  fmt.Sprintf("Invalid trace ID format: %v", err),
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trace, err := th.tr.GetTraceByID(id)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to fetch trace with ID %s for deletion: %v", idStr, err),
		})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// Debug: Log fetched trace
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Fetched trace %s for deletion: bucket_path=%s", idStr, trace.BucketPath),
	})

	if trace.BucketPath != "" {
		parts := strings.Split(trace.BucketPath, "/")
		if len(parts) > 2 {
			objectName := parts[len(parts)-1]
			bucketName := th.config.BucketName

			// Debug: Log GCS deletion attempt
			th.logger.Log(logging.Entry{
				Severity: logging.Debug,
				Payload:  fmt.Sprintf("Attempting to delete object %s from bucket %s", objectName, bucketName),
			})
			err = th.client.Bucket(bucketName).Object(objectName).Delete(th.ctx)
			if err != nil {
				th.logger.Log(logging.Entry{
					Severity: logging.Warning,
					Payload:  fmt.Sprintf("Failed to delete object %s from bucket %s: %v", objectName, bucketName, err),
				})
				// Continue even if GCS deletion fails
			} else {
				th.logger.Log(logging.Entry{
					Severity: logging.Info,
					Payload:  fmt.Sprintf("Deleted object %s from bucket %s", objectName, bucketName),
				})
			}
		}
	}

	// Debug: Log database deletion attempt
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Deleting trace %s from database", idStr),
	})
	err = th.tr.DeleteTrace(id)
	if err != nil {
		th.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload:  fmt.Sprintf("Failed to delete trace with ID %s from database: %v", idStr, err),
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	th.logger.Log(logging.Entry{
		Severity: logging.Info,
		Payload:  fmt.Sprintf("Successfully deleted trace with ID: %s", idStr),
	})
	// Debug: Log response completion
	th.logger.Log(logging.Entry{
		Severity: logging.Debug,
		Payload:  fmt.Sprintf("Completed deletion of trace %s, returning no content", idStr),
	})
	w.WriteHeader(http.StatusNoContent)
}

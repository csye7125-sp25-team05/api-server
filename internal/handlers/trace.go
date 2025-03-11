package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"api-server/internal/model"
	"api-server/internal/repository"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TraceHandler struct {
	tr *repository.TraceRepository
}

func NewTraceHandler(db *sql.DB) *TraceHandler {
	return &TraceHandler{tr: repository.NewTraceRepository(db)}
}

func (th *TraceHandler) GetTraces(w http.ResponseWriter, r *http.Request) {
	traces, err := th.tr.GetAllTraces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(traces)
}

func (th *TraceHandler) GetTraceByID(w http.ResponseWriter, r *http.Request) {
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
	var trace model.Trace
	err := json.NewDecoder(r.Body).Decode(&trace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = th.tr.CreateTrace(&trace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (th *TraceHandler) UpdateTrace(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = th.tr.DeleteTrace(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

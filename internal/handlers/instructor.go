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

type InstructorHandler struct {
	ir *repository.InstructorRepository
}

func NewInstructorHandler(db *sql.DB) *InstructorHandler {
	return &InstructorHandler{ir: repository.NewInstructorRepository(db)}
}

func (ih *InstructorHandler) GetInstructors(w http.ResponseWriter, r *http.Request) {
	instructors, err := ih.ir.GetAllInstructors()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(instructors)
}

func (ih *InstructorHandler) GetInstructorByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	instructor, err := ih.ir.GetInstructorByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(instructor)
}

func (ih *InstructorHandler) CreateInstructor(w http.ResponseWriter, r *http.Request) {
	var instructor model.Instructor
	err := json.NewDecoder(r.Body).Decode(&instructor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ih.ir.CreateInstructor(&instructor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (ih *InstructorHandler) UpdateInstructor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var instructor model.Instructor
	err = json.NewDecoder(r.Body).Decode(&instructor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ih.ir.UpdateInstructor(id, &instructor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (ih *InstructorHandler) DeleteInstructor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ih.ir.DeleteInstructor(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

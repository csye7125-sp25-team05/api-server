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

type CourseHandler struct {
	cr *repository.CourseRepository
}

func NewCourseHandler(db *sql.DB) *CourseHandler {
	return &CourseHandler{cr: repository.NewCourseRepository(db)}
}

func (ch *CourseHandler) GetCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := ch.cr.GetAllCourses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(courses)
}

func (ch *CourseHandler) GetCourseByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	course, err := ch.cr.GetCourseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(course)
}

func (ch *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var course model.Course
	err := json.NewDecoder(r.Body).Decode(&course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ch.cr.CreateCourse(&course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (ch *CourseHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var course model.Course
	err = json.NewDecoder(r.Body).Decode(&course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ch.cr.UpdateCourse(id, &course)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (ch *CourseHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ch.cr.DeleteCourse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

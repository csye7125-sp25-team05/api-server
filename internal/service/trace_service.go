package service

import (
	"api-server/internal/model"
	"api-server/internal/repository"
	"database/sql"

	"github.com/google/uuid"
)

type TraceService struct {
	tr *repository.TraceRepository
}

func NewTraceService(db *sql.DB) *TraceService {
	return &TraceService{tr: repository.NewTraceRepository(db)}
}

func (ts *TraceService) GetAllTraces() ([]model.Trace, error) {
	return ts.tr.GetAllTraces()
}

func (ts *TraceService) GetTraceByID(id uuid.UUID) (*model.Trace, error) {
	return ts.tr.GetTraceByID(id)
}

func (ts *TraceService) CreateTrace(trace *model.Trace) error {
	return ts.tr.CreateTrace(trace)
}

func (ts *TraceService) UpdateTrace(id uuid.UUID, trace *model.Trace) error {
	return ts.tr.UpdateTrace(id, trace)
}

func (ts *TraceService) DeleteTrace(id uuid.UUID) error {
	return ts.tr.DeleteTrace(id)
}

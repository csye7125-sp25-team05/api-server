package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"api-server/internal/model"

	"github.com/google/uuid"
)

type TraceRepository struct {
	db *sql.DB
}

func NewTraceRepository(db *sql.DB) *TraceRepository {
	return &TraceRepository{db: db}
}

func (tr *TraceRepository) GetAllTraces() ([]model.Trace, error) {
	rows, err := tr.db.Query(`SELECT * FROM "api"."trace"`)
	if err != nil {
		fmt.Println("ERROR", err)
		return nil, err
	}
	defer rows.Close()

	var traces []model.Trace
	for rows.Next() {
		var trace model.Trace
		err = rows.Scan(&trace.ID, &trace.UserID, &trace.FileName, &trace.DateCreated, &trace.BucketPath)
		if err != nil {
			return nil, err
		}
		traces = append(traces, trace)
	}
	return traces, nil
}

func (tr *TraceRepository) GetTraceByID(id uuid.UUID) (*model.Trace, error) {
	row := tr.db.QueryRow("SELECT * FROM api.trace WHERE id = $1", id)
	var trace model.Trace
	err := row.Scan(&trace.ID, &trace.UserID, &trace.FileName, &trace.DateCreated, &trace.BucketPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("trace not found")
		}
		return nil, err
	}
	return &trace, nil
}

func (tr *TraceRepository) CreateTrace(trace *model.Trace) error {
	if trace.ID == uuid.Nil {
		trace.ID = uuid.New()
	}
	_, err := tr.db.Exec("INSERT INTO api.trace (id, user_id, file_name, date_created, bucket_path) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4)",
		trace.ID, trace.UserID, trace.FileName, trace.BucketPath)
	return err
}

func (tr *TraceRepository) UpdateTrace(id uuid.UUID, trace *model.Trace) error {
	_, err := tr.db.Exec("UPDATE api.trace SET user_id = $1, file_name = $2, bucket_path = $3 WHERE id = $4",
		trace.UserID, trace.FileName, trace.BucketPath, id)
	return err
}

func (tr *TraceRepository) DeleteTrace(id uuid.UUID) error {
	_, err := tr.db.Exec("DELETE FROM api.trace WHERE id = $1", id)
	return err
}

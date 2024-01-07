package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"newsWebApp/app/newsService/internal/models"
)

type SourceStorage struct {
	db *sql.DB
}

func NewSourceStorage(db *sql.DB) *SourceStorage {
	return &SourceStorage{db: db}
}

func (s *SourceStorage) GetList(ctx context.Context) ([]models.Source, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT source_id, name, feed_url FROM sources")
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	sources := []models.Source{}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoSources
		}
		return nil, fmt.Errorf("can't get sources: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		sour := models.Source{}
		err = rows.Scan(&sour.ID, &sour.Name, &sour.FeedURL)
		if err != nil {
			return nil, fmt.Errorf("can't scan model source: %w", err)
		}

		sources = append(sources, sour)
	}

	return sources, nil
}

func (s *SourceStorage) GetByID(ctx context.Context, id int64) (*models.Source, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT source_id, name, feed_url FROM sources WHERE source_id = $1")
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSourceNotFound
		}
		return nil, fmt.Errorf("can't get source: %w", err)
	}

	source := models.Source{}

	if err := row.Scan(&source.ID, &source.Name, &source.FeedURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSourceNotFound
		}
		return nil, fmt.Errorf("can't scan source: %w", err)
	}

	return &source, nil
}

func (s *SourceStorage) Add(ctx context.Context, source models.Source) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO sources (name, feed_url) VALUES ($1, $2) RETURNING source_id")
	if err != nil {
		return 0, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, source)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSourceExists
		}
		return 0, fmt.Errorf("can't insert source: %w", err)
	}

	var id int64

	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSourceExists
		}
		return 0, fmt.Errorf("can't get last insert id: %w", err)
	}

	return id, nil
}

func (s *SourceStorage) Delete(ctx context.Context, id int64) error {
	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM sources WHERE source_id = $1")
	if err != nil {
		return fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, id); err != nil {
		return fmt.Errorf("can't delete source from db: %v", err)
	}

	return nil
}

package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"newsWebApp/app/newsService/internal/models"
	"newsWebApp/app/newsService/internal/storage"
)

func (s *Storage) SourcesList(ctx context.Context) ([]models.Source, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT source_id, name, feed_url FROM sources")
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	sources := []models.Source{}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSources
		}
		return nil, fmt.Errorf("can't get sources: %w", err)
	}

	for rows.Next() {
		sour := models.Source{}
		err = rows.Scan(&sour.ID, &sour.Name, &sour.FeedURL)
		if err != nil {
			return nil, fmt.Errorf("can't scan model: %w", err)
		}

		sources = append(sources, sour)
	}

	return sources, nil
}

func (s *Storage) SourseByID(ctx context.Context, id int64) (*models.Source, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT source_id, name, feed_url FROM sources WHERE source_id = $1")
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrSourceNotFound
		}
		return nil, fmt.Errorf("can't get source: %w", err)
	}

	sour := models.Source{}

	if err := row.Scan(&sour.ID, &sour.Name, &sour.FeedURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrSourceNotFound
		}
		return nil, fmt.Errorf("can't scan source: %w", err)
	}

	return &sour, nil
}

package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
	"eventlineup/internal/domain/repository"
)

type performanceRepo struct{ pool *pgxpool.Pool }

func NewPerformanceRepository(pool *pgxpool.Pool) *performanceRepo {
	return &performanceRepo{pool: pool}
}

// DJPerformance aggregates one DJ's reps across the organizer's events and adds a
// by-genre breakdown. A DJ that isn't the organizer's yields no aggregate row and
// maps to apperrors.ErrNotFound (mirrors the 404-not-403 scoping convention).
func (r *performanceRepo) DJPerformance(ctx context.Context, djID, organizerID string) (model.DJPerformance, error) {
	var p model.DJPerformance
	err := r.pool.QueryRow(ctx, queryDJPerformance, djID, organizerID).
		Scan(&p.DjID, &p.DjName, &p.Reps, &p.TotalMinutes, &p.LastPlayed)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.DJPerformance{}, apperrors.ErrNotFound
	}
	if err != nil {
		return model.DJPerformance{}, err
	}

	rows, err := r.pool.Query(ctx, queryDJPerformanceByGenre, djID, organizerID)
	if err != nil {
		return model.DJPerformance{}, err
	}
	defer rows.Close()
	p.ByGenre = []model.GenreStat{}
	for rows.Next() {
		var g model.GenreStat
		if err := rows.Scan(&g.Genre, &g.Reps, &g.TotalMinutes); err != nil {
			return model.DJPerformance{}, err
		}
		p.ByGenre = append(p.ByGenre, g)
	}
	return p, rows.Err()
}

// RosterSummary returns every active student of the organizer with their reps in
// the window, including zero-rep students.
func (r *performanceRepo) RosterSummary(ctx context.Context, organizerID string, filter repository.PerformanceFilter) ([]model.RosterPerformance, error) {
	rows, err := r.pool.Query(ctx, queryRosterSummary, organizerID, filter.EventID, filter.From, filter.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.RosterPerformance{}
	for rows.Next() {
		var rp model.RosterPerformance
		if err := rows.Scan(&rp.DjID, &rp.DjName, &rp.IsStudent, &rp.Reps, &rp.TotalMinutes, &rp.LastPlayed); err != nil {
			return nil, err
		}
		out = append(out, rp)
	}
	return out, rows.Err()
}

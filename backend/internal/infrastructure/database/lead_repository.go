package database

import (
	"context"

	"eventlineup/internal/domain/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type leadRepo struct{ pool *pgxpool.Pool }

func NewLeadRepository(pool *pgxpool.Pool) *leadRepo { return &leadRepo{pool: pool} }

func (r *leadRepo) Create(ctx context.Context, l model.Lead) (model.Lead, error) {
	var out model.Lead
	err := r.pool.QueryRow(ctx, queryLeadInsert, l.Name, l.Organization, l.Email, l.Message, l.Source).
		Scan(&out.ID, &out.Name, &out.Organization, &out.Email, &out.Message, &out.Source, &out.CreatedAt)
	return out, err
}

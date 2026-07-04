package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/domain/model"
)

type membershipRepo struct{ pool *pgxpool.Pool }

func NewMembershipRepository(pool *pgxpool.Pool) *membershipRepo {
	return &membershipRepo{pool: pool}
}

func (r *membershipRepo) Create(ctx context.Context, schoolID string, organizerID *string, email, role, status string) (model.Membership, error) {
	var m model.Membership
	err := r.pool.QueryRow(ctx, queryMembershipInsert, schoolID, organizerID, email, role, status).
		Scan(&m.ID, &m.SchoolID, &m.OrganizerID, &m.Email, &m.Role, &m.Status, &m.CreatedAt)
	return m, err
}

// ActiveForOrganizer returns the organizer's active membership (their current
// school), or apperrors.ErrNotFound when they haven't joined one yet.
func (r *membershipRepo) ActiveForOrganizer(ctx context.Context, organizerID string) (model.Membership, error) {
	var m model.Membership
	err := r.pool.QueryRow(ctx, queryMembershipActiveForOrganizer, organizerID).
		Scan(&m.ID, &m.SchoolID, &m.OrganizerID, &m.Email, &m.Role, &m.Status, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Membership{}, apperrors.ErrNotFound
	}
	return m, err
}

func (r *membershipRepo) ListBySchool(ctx context.Context, schoolID string) ([]model.Membership, error) {
	rows, err := r.pool.Query(ctx, queryMembershipListBySchool, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memberships := []model.Membership{}
	for rows.Next() {
		var m model.Membership
		if err := rows.Scan(&m.ID, &m.SchoolID, &m.OrganizerID, &m.Email, &m.Role, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

// LinkPendingByEmail attaches a pending membership to organizerID on their first
// matching sign-in and activates it. A no-op when nothing is pending for email.
func (r *membershipRepo) LinkPendingByEmail(ctx context.Context, organizerID, email string) error {
	_, err := r.pool.Exec(ctx, queryMembershipLinkPendingByEmail, organizerID, email)
	return err
}

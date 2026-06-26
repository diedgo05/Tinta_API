// Package postgres implements ports.MemberRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/community/internal/member/domain"
)

type MemberRepository struct{ db *pgxpool.Pool }

func NewMemberRepository(db *pgxpool.Pool) *MemberRepository { return &MemberRepository{db: db} }

type row struct {
	ID       uuid.UUID
	ClubID   uuid.UUID
	UserID   uuid.UUID
	Role     string
	JoinedAt time.Time
}

func (r row) toDomain() *domain.ClubMember {
	return &domain.ClubMember{
		ID: r.ID, ClubID: r.ClubID, UserID: r.UserID,
		Role: domain.Role(r.Role), JoinedAt: r.JoinedAt,
	}
}

const cols = `id, club_id, user_id, role, joined_at`

func (r *MemberRepository) Join(ctx context.Context, clubID, userID uuid.UUID, role domain.Role) (*domain.ClubMember, error) {
	const q = `INSERT INTO club_members (club_id, user_id, role) VALUES ($1,$2,$3) RETURNING ` + cols
	var rr row
	err := r.db.QueryRow(ctx, q, clubID, userID, string(role)).Scan(
		&rr.ID, &rr.ClubID, &rr.UserID, &rr.Role, &rr.JoinedAt)
	if err != nil {
		if strings.Contains(err.Error(), "club_members_club_id_user_id_key") {
			return nil, domain.ErrAlreadyMember
		}
		return nil, fmt.Errorf("insert member: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *MemberRepository) Leave(ctx context.Context, clubID, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM club_members WHERE club_id=$1 AND user_id=$2`, clubID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	return nil
}

func (r *MemberRepository) GetMembership(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubMember, error) {
	const q = `SELECT ` + cols + ` FROM club_members WHERE club_id=$1 AND user_id=$2`
	var rr row
	err := r.db.QueryRow(ctx, q, clubID, userID).Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.Role, &rr.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *MemberRepository) ListByClub(ctx context.Context, clubID uuid.UUID) ([]*domain.ClubMember, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cols+` FROM club_members WHERE club_id=$1 ORDER BY joined_at ASC`, clubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.ClubMember, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.Role, &rr.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *MemberRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.ClubMember, error) {
	rows, err := r.db.Query(ctx, `SELECT `+cols+` FROM club_members WHERE user_id=$1 ORDER BY joined_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.ClubMember, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.Role, &rr.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *MemberRepository) CountMembers(ctx context.Context, clubID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM club_members WHERE club_id=$1`, clubID).Scan(&n)
	return n, err
}

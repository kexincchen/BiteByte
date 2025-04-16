package postgres

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (username, email, password, role, created_at, updated_at)
	          VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	return r.db.QueryRowContext(
		ctx, query,
		u.Username, u.Email, u.Password, u.Role, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
}

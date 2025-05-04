package postgres

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `INSERT INTO users
		(username, email, password, role, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id`

	return r.db.QueryRowContext(ctx, q,
		u.Username, u.Email, u.Password, u.Role, u.CreatedAt, u.UpdatedAt,
	).Scan(&u.ID)
}

func (r *UserRepo) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	const q = `SELECT id, username, email, password, role,
	           created_at, updated_at
	           FROM users WHERE id=$1`
	var u domain.User
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `SELECT id, username, email, password, role,
	           created_at, updated_at
	           FROM users WHERE email=$1`
	var u domain.User
	if err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	const q = `UPDATE users
		SET username=$1, email=$2, password=$3, role=$4, updated_at=$5
		WHERE id=$6`
	_, err := r.db.ExecContext(ctx, q,
		u.Username, u.Email, u.Password, u.Role, u.UpdatedAt, u.ID,
	)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

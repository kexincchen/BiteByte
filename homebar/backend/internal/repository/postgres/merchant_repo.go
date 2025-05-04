package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kexincchen/homebar/internal/domain"
)

type MerchantRepo struct{ db *sql.DB }

func NewMerchantRepository(db *sql.DB) *MerchantRepo { return &MerchantRepo{db} }

func (r *MerchantRepo) Create(ctx context.Context, m *domain.Merchant) error {
	const q = `INSERT INTO merchants
	  (user_id, business_name, description, address, phone,
	   username, is_verified, created_at, updated_at)
	  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`
	return r.db.QueryRowContext(ctx, q,
		m.UserID, m.BusinessName, m.Description, m.Address, m.Phone,
		m.Username, m.IsVerified, m.CreatedAt, m.UpdatedAt,
	).Scan(&m.ID)
}

func (r *MerchantRepo) GetByID(ctx context.Context, id uint) (*domain.Merchant, error) {
	const q = `SELECT id, user_id, business_name, description, address,
	           phone, username, is_verified
	           FROM merchants WHERE id=$1`
	var m domain.Merchant
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&m.ID, &m.UserID, &m.BusinessName, &m.Description, &m.Address,
		&m.Phone, &m.Username, &m.IsVerified,
	); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MerchantRepo) GetByUsername(ctx context.Context, u string) (*domain.Merchant, error) {
	const q = `SELECT id, user_id, business_name, description, address,
	           phone, username, is_verified
	           FROM merchants WHERE username=$1`
	var m domain.Merchant
	if err := r.db.QueryRowContext(ctx, q, u).Scan(
		&m.ID, &m.UserID, &m.BusinessName, &m.Description, &m.Address,
		&m.Phone, &m.Username, &m.IsVerified,
	); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MerchantRepo) List(ctx context.Context) ([]*domain.Merchant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, business_name, description, address,
		        phone, username, is_verified
		   FROM merchants ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	var list []*domain.Merchant
	for rows.Next() {
		var m domain.Merchant
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.BusinessName, &m.Description, &m.Address,
			&m.Phone, &m.Username, &m.IsVerified,
		); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *MerchantRepo) Update(ctx context.Context, m *domain.Merchant) error {
	const q = `UPDATE merchants
		SET business_name=$1, description=$2, address=$3, phone=$4,
		    username=$5, is_verified=$6, updated_at=$7
		WHERE id=$8`
	_, err := r.db.ExecContext(ctx, q,
		m.BusinessName, m.Description, m.Address, m.Phone,
		m.Username, m.IsVerified, m.UpdatedAt, m.ID,
	)
	return err
}

func (r *MerchantRepo) Delete(ctx context.Context, id uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM merchants WHERE id=$1`, id)
	return err
}

func (r *MerchantRepo) GetByUserID(ctx context.Context, userID uint) (*domain.Merchant, error) {
	const q = `SELECT id, user_id, business_name, description, address, phone, username,
			   is_verified, created_at, updated_at
			   FROM merchants WHERE user_id=$1`

	var m domain.Merchant
	err := r.db.QueryRowContext(ctx, q, userID).Scan(
		&m.ID, &m.UserID, &m.BusinessName, &m.Description, &m.Address, &m.Phone, &m.Username,
		&m.IsVerified, &m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no merchant found for user ID %d", userID)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &m, nil
}

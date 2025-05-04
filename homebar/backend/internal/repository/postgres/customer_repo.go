package postgres

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/domain"
)

type CustomerRepo struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) Create(ctx context.Context, c *domain.Customer) error {
	const q = `INSERT INTO customers
		(user_id, first_name, last_name, address, phone)
		VALUES ($1,$2,$3,$4,$5)`

	_, err := r.db.ExecContext(ctx, q,
		c.UserID, c.FirstName, c.LastName, c.Address, c.Phone,
	)
	return err
}

func (r *CustomerRepo) GetByUserID(ctx context.Context, userID uint) (*domain.Customer, error) {
	const q = `SELECT user_id, first_name, last_name, address, phone
	           FROM customers WHERE user_id=$1`
	var c domain.Customer
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(
		&c.UserID, &c.FirstName, &c.LastName, &c.Address, &c.Phone,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepo) Update(ctx context.Context, c *domain.Customer) error {
	const q = `UPDATE customers
		SET first_name=$1, last_name=$2, address=$3, phone=$4
		WHERE user_id=$5`
	_, err := r.db.ExecContext(ctx, q,
		c.FirstName, c.LastName, c.Address, c.Phone, c.UserID,
	)
	return err
}

func (r *CustomerRepo) Delete(ctx context.Context, userID uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM customers WHERE user_id=$1`, userID)
	return err
} 
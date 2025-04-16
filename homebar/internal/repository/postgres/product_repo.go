package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kexincchen/homebar/internal/domain"
)

// ProductRepo is the Postgres implementation of repository.ProductRepository.
type ProductRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepo { return &ProductRepo{db: db} }

// -----------------------  CRUD  -----------------------

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	const q = `INSERT INTO products
		(merchant_id, name, description, price, category, image_url, is_available,
		 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`

	return r.db.QueryRowContext(ctx, q,
		p.MerchantID, p.Name, p.Description, p.Price, p.Category,
		p.ImageURL, p.IsAvailable, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
}

func (r *ProductRepo) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	const q = `SELECT id, merchant_id, name, description, price, category,
			   image_url, is_available, created_at, updated_at
			   FROM products WHERE id=$1`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Category,
		&p.ImageURL, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Product, error) {
	q := `SELECT id, merchant_id, name, description, price, category,
		  image_url, is_available, created_at, updated_at
		  FROM products`
	var rows *sql.Rows
	var err error

	if merchantID == 0 {
		rows, err = r.db.QueryContext(ctx, q) // all products
	} else {
		rows, err = r.db.QueryContext(ctx, q+" WHERE merchant_id=$1", merchantID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Category,
			&p.ImageURL, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	const q = `UPDATE products
			   SET name=$1, description=$2, price=$3, category=$4,
				   image_url=$5, is_available=$6, updated_at=$7
			   WHERE id=$8`
	res, err := r.db.ExecContext(ctx, q,
		p.Name, p.Description, p.Price, p.Category,
		p.ImageURL, p.IsAvailable, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("product %d not found", p.ID)
	}
	return nil
}

func (r *ProductRepo) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("product %d not found", id)
	}
	return nil
}

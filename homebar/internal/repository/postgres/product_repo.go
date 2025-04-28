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

// CRUD

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	const q = `INSERT INTO products
		(merchant_id, name, description, price, category, mime_type, image_data, is_available,
		 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id`

	return r.db.QueryRowContext(ctx, q,
		p.MerchantID, p.Name, p.Description, p.Price, p.Category,
		p.MimeType, p.ImageData, p.IsAvailable, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
}

func (r *ProductRepo) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	const q = `SELECT id, merchant_id, name, description, price, category,
			   mime_type, image_data, is_available, created_at, updated_at
			   FROM products WHERE id=$1`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Category,
		&p.MimeType, &p.ImageData, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	var (
		q    string
		args []interface{}
	)

	if len(p.ImageData) == 0 {
		// No new image uploaded: don't update mime_type and image_data
		q = `UPDATE products
			     SET name=$1, description=$2, price=$3, category=$4,
			         is_available=$5, updated_at=$6
			     WHERE id=$7`
		args = []interface{}{
			p.Name, p.Description, p.Price, p.Category,
			p.IsAvailable, p.UpdatedAt, p.ID,
		}
	} else {
		// New image uploaded: update mime_type and image_data
		q = `UPDATE products
			     SET name=$1, description=$2, price=$3, category=$4,
			         mime_type=$5, image_data=$6, is_available=$7, updated_at=$8
			     WHERE id=$9`
		args = []interface{}{
			p.Name, p.Description, p.Price, p.Category,
			p.MimeType, p.ImageData, p.IsAvailable, p.UpdatedAt, p.ID,
		}
	}

	res, err := r.db.ExecContext(ctx, q, args...)
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

func (r *ProductRepo) GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Product, error) {
	q := `SELECT id, merchant_id, name, description, price, category,
		  mime_type, image_data, is_available, created_at, updated_at
		  FROM products`
	var rows *sql.Rows
	var err error

	rows, err = r.db.QueryContext(ctx, q+" WHERE merchant_id=$1", merchantID)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
		}
	}(rows)

	var list []*domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Category,
			&p.MimeType, &p.ImageData, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]*domain.Product, error) {
	const q = `SELECT id, merchant_id, name, description, price, category,
	           mime_type, image_data, is_available, created_at, updated_at
	           FROM products`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
		}
	}(rows)

	var list []*domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Category,
			&p.MimeType, &p.ImageData, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/kexincchen/homebar/internal/domain"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepository(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

// Create -------  Create (order + items in one TX) -------
func (r *OrderRepo) Create(ctx context.Context, o *domain.Order, items []domain.OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	const qOrder = `INSERT INTO orders
	  (customer_id, merchant_id, total_amount, status, notes, created_at, updated_at)
	  VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	if err := tx.QueryRowContext(ctx, qOrder,
		o.CustomerID, o.MerchantID, o.TotalAmount, o.Status, o.Notes,
		o.CreatedAt, o.UpdatedAt,
	).Scan(&o.ID); err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	const qItem = `INSERT INTO order_items
	  (order_id, product_id, quantity, price)
	  VALUES ($1,$2,$3,$4)`
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, qItem, o.ID, it.ProductID, it.Quantity, it.Price); err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}
	return tx.Commit()
}

// -------  Query helpers  -------
func scanOrder(row *sql.Row) (*domain.Order, error) {
	var o domain.Order
	if err := row.Scan(&o.ID, &o.CustomerID, &o.MerchantID, &o.TotalAmount,
		&o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, merchant_id, total_amount, status, notes,
		        created_at, updated_at FROM orders WHERE id=$1`, id)
	o, err := scanOrder(row)
	if err != nil {
		return nil, nil, err
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, product_id, quantity, price
		   FROM order_items WHERE order_id=$1`, id)
	if err != nil {
		return nil, nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)
	var list []domain.OrderItem
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID,
			&it.Quantity, &it.Price); err != nil {
			return nil, nil, err
		}
		list = append(list, it)
	}
	return o, list, nil
}

func (r *OrderRepo) GetByCustomer(ctx context.Context, cid uint) ([]*domain.Order, error) {
	return r.list(ctx, `WHERE customer_id=$1`, cid)
}
func (r *OrderRepo) GetByMerchant(ctx context.Context, mid uint) ([]*domain.Order, error) {
	return r.list(ctx, `WHERE merchant_id=$1`, mid)
}

func (r *OrderRepo) list(ctx context.Context, where string, arg interface{}) ([]*domain.Order, error) {
	q := fmt.Sprintf(`SELECT id, customer_id, merchant_id, total_amount, status, notes,
	                   created_at, updated_at FROM orders %s ORDER BY created_at DESC`, where)
	rows, err := r.db.QueryContext(ctx, q, arg)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)
	var list []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.MerchantID, &o.TotalAmount,
			&o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &o)
	}
	return list, rows.Err()
}

// UpdateStatus updates the status of an order
func (r *OrderRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, id uint, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, status, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}

	return err
}

// Update updates an order
func (r *OrderRepo) Update(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	query := `
		UPDATE orders 
		SET status = $1, notes = $2, updated_at = NOW()
		WHERE id = $3
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, order.Status, order.Notes, order.ID)
	} else {
		_, err = r.db.ExecContext(ctx, query, order.Status, order.Notes, order.ID)
	}

	return err
}

// Delete removes an order and its items from the database
func (r *OrderRepo) Delete(ctx context.Context, tx *sql.Tx, id uint) error {
	// Delete order items first due to foreign key constraints
	_, err := tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return err
	}

	// Then delete the order
	_, err = tx.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	return err
}

// GetDB returns the underlying database connection
func (r *OrderRepo) GetDB() *sql.DB {
	return r.db
}

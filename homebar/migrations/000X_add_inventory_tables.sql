BEGIN;

-- Add status column to inventory_items if it doesn't exist
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'available';

-- Create inventory_reservations table for tracking inventory locks during orders
CREATE TABLE IF NOT EXISTS inventory_reservations (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    ingredient_id INT NOT NULL REFERENCES inventory_items(id),
    quantity DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'reserved', -- reserved, completed, canceled
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_reservations_order_id ON inventory_reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_ingredient_status ON inventory_reservations(ingredient_id, status);

COMMIT; 
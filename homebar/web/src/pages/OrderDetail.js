import React, { useState, useEffect } from 'react';
import { useParams, useLocation } from 'react-router-dom';
import { orderAPI } from '../services/api';

const OrderDetail = () => {
  const [order, setOrder] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  const { id } = useParams();
  const location = useLocation();

  useEffect(() => {
    // If order was passed via navigation state, use it
    if (location.state?.order) {
      setOrder(location.state.order);
      setLoading(false);
      return;
    }

    // Otherwise fetch from API
    const fetchOrder = async () => {
      try {
        const response = await orderAPI.getOrder(id);
        setOrder(response.data);
        setLoading(false);
      } catch (error) {
        console.error('Error fetching order:', error);
        setError('Failed to load order details');
        setLoading(false);
      }
    };

    fetchOrder();
  }, [id, location.state]);

  if (loading) return <div>Loading order details...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!order) return <div className="error">Order not found</div>;

  return (
    <div className="order-detail-page">
      <h1>Order #{order.id}</h1>
      
      <div className="order-meta">
        <div className="order-status">
          Status: <span className={`status-${order.status}`}>{order.status}</span>
        </div>
        <div className="order-date">
          Ordered on: {new Date(order.created_at).toLocaleString()}
        </div>
      </div>
      
      {order.notes && (
        <div className="order-notes">
          <h3>Notes</h3>
          <p>{order.notes}</p>
        </div>
      )}
      
      <div className="order-items">
        <h3>Items</h3>
        <table>
          <thead>
            <tr>
              <th>Item</th>
              <th>Quantity</th>
              <th>Price</th>
              <th>Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {order.items?.map(item => (
              <tr key={item.id}>
                <td>{item.product_name}</td>
                <td>{item.quantity}</td>
                <td>${item.price.toFixed(2)}</td>
                <td>${(item.price * item.quantity).toFixed(2)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td colSpan="3" align="right"><strong>Total:</strong></td>
              <td><strong>${order.total_amount.toFixed(2)}</strong></td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  );
};

export default OrderDetail; 
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { orderAPI } from '../../services/api';

const MerchantOrderDetail = () => {
    const [order, setOrder] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const { id } = useParams();

    useEffect(() => {
        const fetchOrderDetails = async () => {
            try {
                const response = await orderAPI.getOrder(id);
                const { order, items } = response.data;
                setOrder({
                    ...order,
                    items: items
                });
                setLoading(false);
            } catch (error) {
                console.error('Error fetching order:', error);
                setError('Failed to load order details');
                setLoading(false);
            }
        };

        fetchOrderDetails();
    }, [id]);

    const handleStatusChange = async (newStatus) => {
        try {
            await orderAPI.updateOrderStatus(order.id, newStatus);
            setOrder(prev => ({
                ...prev,
                status: newStatus
            }));
        } catch (error) {
            console.error('Error updating order status:', error);
        }
    };

    if (loading) return <div>Loading order details...</div>;
    if (error) return <div className="error">{error}</div>;
    if (!order) return <div className="error">Order not found</div>;

    return (
        <div className="merchant-order-detail-page">
            <div className="order-header">
                <h1>Order #{order.id}</h1>
                <div className="order-meta">
                    <div className="status-section">
                        <label>Status: </label>
                        <select
                            value={order.status}
                            onChange={(e) => handleStatusChange(e.target.value)}
                            className={`status-select status-${order.status}`}
                        >
                            <option value="pending">Pending</option>
                            <option value="confirmed">Confirmed</option>
                            <option value="preparing">Preparing</option>
                            <option value="ready">Ready</option>
                            <option value="delivered">Delivered</option>
                            <option value="cancelled">Cancelled</option>
                            <option value="refunded">Refunded</option>
                        </select>
                    </div>
                    <div className="order-date">
                        Ordered on: {new Date(order.created_at).toLocaleString()}
                    </div>
                    <div className="customer-info">
                        Customer ID: {order.customer_id}
                    </div>
                </div>
            </div>

            {order.notes && (
                <div className="order-notes">
                    <h3>Order Notes</h3>
                    <p>{order.notes}</p>
                </div>
            )}

            <div className="order-items">
                <h3>Order Items</h3>
                <table className="items-table">
                    <thead>
                    <tr>
                        <th>Item</th>
                        <th>Description</th>
                        <th>Quantity</th>
                        <th>Price</th>
                        <th>Subtotal</th>
                    </tr>
                    </thead>
                    <tbody>
                    {order.items.map(item => (
                        <tr key={item.id}>
                            <td>{item.product_name}</td>
                            <td>{item.product_description}</td>
                            <td>{item.quantity}</td>
                            <td>${item.price.toFixed(2)}</td>
                            <td>${(item.price * item.quantity).toFixed(2)}</td>
                        </tr>
                    ))}
                    </tbody>
                    <tfoot>
                    <tr>
                        <td colSpan="4" align="right"><strong>Total:</strong></td>
                        <td><strong>${order.total_amount.toFixed(2)}</strong></td>
                    </tr>
                    </tfoot>
                </table>
            </div>
        </div>
    );
};

export default MerchantOrderDetail;

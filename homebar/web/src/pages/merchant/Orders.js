import React, { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import { orderAPI } from "../../services/api";
import { AuthContext } from "../../contexts/AuthContext";

const Orders = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { currentUser } = useContext(AuthContext);

  useEffect(() => {
    const fetchOrders = async () => {
      if (!currentUser || !currentUser.merchant_id) {
        setError("You must be logged in as a merchant to view this page");
        setLoading(false);
        return;
      }

      try {
        const response = await orderAPI.getOrdersByMerchant(
          currentUser.merchant_id
        );
        setOrders(Array.isArray(response.data) ? response.data : []);
        setLoading(false);
      } catch (err) {
        console.error("Error fetching merchant orders:", err);
        setError("Failed to load merchant orders");
        setLoading(false);
      }
    };

    fetchOrders();
  }, [currentUser]);

  if (loading) return <div>Loading orders...</div>;
  if (error) return <div className="error">{error}</div>;
  if (orders.length === 0)
    return <div>You haven't received any orders yet</div>;

  return (
    <div className="merchant-orders-page">
      <h1>Order Management</h1>

      <table className="orders-table">
        <thead>
          <tr>
            <th>Order ID</th>
            <th>Customer</th>
            <th>Date</th>
            <th>Status</th>
            <th>Total</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => (
            <tr key={order.id}>
              <td>#{order.id}</td>
              <td>Customer #{order.customer_id}</td>
              <td>{new Date(order.created_at).toLocaleDateString()}</td>
              <td>
                <span className={`status-badge status-${order.status}`}>
                  {order.status}
                </span>
              </td>
              <td>${order.total_amount.toFixed(2)}</td>
              <td>
                <Link to={`/merchant/orders/${order.id}`} className="view-btn">
                  View Details
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Orders;

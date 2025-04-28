import React, { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import { orderAPI } from "../services/api";
import { AuthContext } from "../contexts/AuthContext";

const OrderHistory = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { currentUser } = useContext(AuthContext);

  useEffect(() => {
    const fetchOrders = async () => {
      console.log("Current user: ", currentUser);
      if (!currentUser || !currentUser.id) {
        setLoading(false);
        return;
      }

      try {
        const response = await orderAPI.getOrdersByCustomer(
          currentUser.id
        );
        console.log("Orders: ", response);
        setOrders(Array.isArray(response.data) ? response.data : []);
        setLoading(false);
      } catch (error) {
        console.error("Error fetching orders:", error);
        setError("Failed to load your orders");
        setLoading(false);
      }
    };

    fetchOrders();
  }, [currentUser]);

  if (loading) return <div>Loading orders...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!currentUser) return <div>Please login to view your orders</div>;
  if (orders.length === 0) return <div>You have no orders yet</div>;

  return (
    <div className="order-history-page">
      <h1>Your Orders</h1>

      <div className="orders-list">
        {orders.map((order) => (
          <div key={order.id} className="order-card">
            <div className="order-header">
              <div className="order-id">
                <h3>Order #{order.id}</h3>
                <span className={`order-status status-${order.status}`}>
                  {order.status}
                </span>
              </div>
              <div className="order-date">
                {new Date(order.created_at).toLocaleDateString()}
              </div>
            </div>

            <div className="order-details">
              <div className="order-total">
                Total: ${order.total_amount.toFixed(2)}
              </div>

              <Link
                to={`/orders/${order.id}`}
                state={{ order }}
                className="view-order-button"
              >
                View Details
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default OrderHistory;

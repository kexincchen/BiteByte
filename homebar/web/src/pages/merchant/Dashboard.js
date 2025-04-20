import React, { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import { orderAPI, productAPI } from "../../services/api";
import { AuthContext } from "../../contexts/AuthContext";

const Dashboard = () => {
  const { currentUser } = useContext(AuthContext);
  const [recentOrders, setRecentOrders] = useState([]);
  const [productCount, setProductCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchDashboardData = async () => {
      if (!currentUser || currentUser.role !== "merchant") {
        setError("You must be logged in as a merchant to view this page");
        setLoading(false);
        return;
      }

      try {
        // First, make sure merchant_id exists and is valid
        if (!currentUser.merchant_id) {
          console.error("No merchant_id found for the current user");
          setError("Your merchant account is not properly set up. Please contact support.");
          setLoading(false);
          return;
        }

        // Fetch recent orders - include merchant_id as a query parameter
        const ordersResponse = await orderAPI.getOrdersByMerchant(currentUser.merchant_id);
        setRecentOrders(Array.isArray(ordersResponse.data) 
          ? ordersResponse.data.slice(0, 5) 
          : []);

        // Fetch products count
        const productsResponse = await productAPI.getProductsByMerchant(currentUser.merchant_id);
        setProductCount(Array.isArray(productsResponse.data) 
          ? productsResponse.data.length 
          : 0);

        setLoading(false);
      } catch (error) {
        console.error("Error fetching dashboard data:", error);
        // More detailed error message
        let errorMessage = "Failed to load dashboard data";
        if (error.response) {
          if (error.response.status === 400) {
            errorMessage += ": Invalid request. Please check your merchant profile.";
          } else if (error.response.status === 401) {
            errorMessage += ": Authentication error. Please log in again.";
          } else if (error.response.status === 403) {
            errorMessage += ": You don't have permission to access this data.";
          }
        }
        
        setError(errorMessage);
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, [currentUser]);

  if (loading) return <div>Loading dashboard...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="merchant-dashboard">
      <h1>Merchant Dashboard</h1>

      <div className="dashboard-stats">
        <div className="stat-card">
          <h3>Products</h3>
          <p className="stat-number">{productCount}</p>
          <Link to="/merchant/products" className="action-link">
            Manage Menu
          </Link>
        </div>
        <div className="stat-card">
          <h3>Recent Orders</h3>
          <p className="stat-number">{recentOrders.length}</p>
          <Link to="/merchant/orders" className="action-link">
            View All Orders
          </Link>
        </div>
        <div className="stat-card">
          <h3>Quick Actions</h3>
          <div className="action-buttons">
            <Link to="/merchant/products/new" className="action-button">
              Add Product
            </Link>
            <Link to="/merchant/inventory" className="action-button">
              Manage Inventory
            </Link>
          </div>
        </div>
      </div>

      <div className="recent-orders-section">
        <h2>Recent Orders</h2>
        {recentOrders.length === 0 ? (
          <p>No recent orders</p>
        ) : (
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
              {recentOrders.map((order) => (
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
                    <Link to={`/merchant/orders/${order.id}`}>Edit</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default Dashboard;

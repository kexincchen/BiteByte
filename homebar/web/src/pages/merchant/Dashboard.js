import React, { useState, useEffect, useContext } from "react";
import { Link } from "react-router-dom";
import { orderAPI, productAPI, ingredientAPI } from "../../services/api";
import { AuthContext } from "../../contexts/AuthContext";

const Dashboard = () => {
  const { currentUser, updateCurrentUser } = useContext(AuthContext);
  const [recentOrders, setRecentOrders] = useState([]);
  const [productCount, setProductCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [ingredientStats, setIngredientStats] = useState({
    total: 0,
    lowStock: 0,
  });

  useEffect(() => {
    const fetchDashboardData = async () => {
      if (!currentUser || currentUser.role !== "merchant") {
        setError("You must be logged in as a merchant to view this page");
        setLoading(false);
        return;
      }

      try {
        console.log("Current user: ", currentUser);

        if (currentUser.merchant_id) {
          console.log("Fetching merchant data for merchant_id: ", currentUser.merchant_id);
          // If merchant_id exists, use it directly
          await fetchMerchantData(currentUser.merchant_id);
        } else {
          // Otherwise, try to fetch the merchant data by user ID
          try {
            const merchantResponse = await fetch(
              `/api/merchants/user/${currentUser.id}`,
              {
                headers: {
                  Authorization: `Bearer ${localStorage.getItem("token")}`,
                },
              }
            );

            if (!merchantResponse.ok) {
              throw new Error(`HTTP error! Status: ${merchantResponse.status}`);
            }

            const merchantData = await merchantResponse.json();
            console.log("Merchant data: ", merchantData);

            if (merchantData && merchantData.id) {
              // Update the user with merchant_id
              const updatedUser = {
                ...currentUser,
                merchant_id: merchantData.id,
                business_name:
                  merchantData.business_name || merchantData.businessName,
              };
              updateCurrentUser(updatedUser);

              // Continue with the merchant data
              await fetchMerchantData(merchantData.id);
            } else {
              setError(
                "Could not retrieve merchant information. Please contact support."
              );
              setLoading(false);
            }
          } catch (err) {
            console.error("Error fetching merchant data:", err);
            setError("Failed to load merchant profile: " + err.message);
            setLoading(false);
          }
        }
      } catch (err) {
        console.error("Dashboard initialization error:", err);
        setError("An error occurred while loading dashboard data");
        setLoading(false);
      }
    };

    const fetchMerchantData = async (merchantId) => {
      try {
        // Get orders
        const ordersResponse = await orderAPI.getOrdersByMerchant(merchantId);
        setRecentOrders(
          Array.isArray(ordersResponse.data)
            ? ordersResponse.data.slice(0, 5)
            : []
        );

        // Get product count
        const productsResponse = await productAPI.getProductsByMerchant(
          merchantId
        );
        setProductCount(
          Array.isArray(productsResponse.data)
            ? productsResponse.data.length
            : 0
        );

        // Get ingredient inventory stats
        try {
          const inventoryResponse = await ingredientAPI.getInventorySummary(merchantId);
          console.log("Inventory response: ", inventoryResponse);
          if (inventoryResponse.ok) {
            const inventoryData = await inventoryResponse.json();
            console.log("Inventory response: ", inventoryData);
            setIngredientStats({
              total: inventoryData.totalIngredients || 0,
              lowStock: inventoryData.lowStockCount || 0,
            });
          }
        } catch (inventoryError) {
          console.error("Error fetching inventory data:", inventoryError);
          // Don't fail the entire dashboard if just inventory fails
        }

        setLoading(false);
      } catch (error) {
        console.error("Error fetching merchant data:", error);
        setError("Failed to load merchant data: " + error.message);
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, [currentUser, updateCurrentUser]);

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
          <h3>Ingredients</h3>
          <p className="stat-number">{ingredientStats.total}</p>
          {ingredientStats.lowStock > 0 && (
            <p className="warning-text">
              {ingredientStats.lowStock} ingredients low on stock
            </p>
          )}
          <Link to="/merchant/inventory" className="action-link">
            Manage Inventory
          </Link>
        </div>
        <div className="stat-card">
          <h3>Quick Actions</h3>
          <div className="action-buttons">
            <Link to="/merchant/products/new" className="action-button">
              Add Product
            </Link>
            <Link to="/merchant/inventory/add" className="action-button">
              Add Ingredients
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

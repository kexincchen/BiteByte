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
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [showOrderModal, setShowOrderModal] = useState(false);
  const [updatingStatus, setUpdatingStatus] = useState(false);

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
          console.log(
            "Fetching merchant data for merchant_id: ",
            currentUser.merchant_id
          );
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
          const inventoryResponse = await ingredientAPI.getInventorySummary(
            merchantId
          );
          console.log("Inventory response: ", inventoryResponse);
          if (
            inventoryResponse.status >= 200 &&
            inventoryResponse.status < 300
          ) {
            const inventoryData = await inventoryResponse.data;
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

  const handleStatusChange = async (orderId, newStatus) => {
    setUpdatingStatus(true);
    try {
      const response = await orderAPI.updateOrderStatus(orderId, newStatus);
      console.log("Response: ", response);
      if (response.status >= 200 && response.status < 300) {
        // Update the local state to reflect the change
        setRecentOrders(
          recentOrders.map((order) =>
            order.id === orderId ? { ...order, status: newStatus } : order
          )
        );
      } else {
        throw new Error("Failed to update order status");
      }
    } catch (error) {
      console.error("Error updating order status:", error);
      setError(
        "Failed to update order status: " + (error.message || "Unknown error")
      );
    } finally {
      setUpdatingStatus(false);
    }
  };

  const openOrderModal = (order) => {
    setSelectedOrder(order);
    setShowOrderModal(true);
  };

  const closeOrderModal = () => {
    setSelectedOrder(null);
    setShowOrderModal(false);
  };

  const updateOrder = async (updatedOrder) => {
    try {
      const response = await orderAPI.updateOrder(
        updatedOrder.id,
        updatedOrder
      );
      if (response.status >= 200 && response.status < 300) {
        // Update the local state with the updated order
        setRecentOrders(
          recentOrders.map((order) =>
            order.id === updatedOrder.id ? updatedOrder : order
          )
        );
        closeOrderModal();
      } else {
        throw new Error("Failed to update order");
      }
    } catch (error) {
      console.error("Error updating order:", error);
      setError("Failed to update order: " + (error.message || "Unknown error"));
    }
  };

  const handleDeleteOrder = async (orderId) => {
    if (
      !window.confirm(
        "Are you sure you want to delete this order? This action cannot be undone."
      )
    ) {
      return;
    }

    setUpdatingStatus(true);
    try {
      // First set status to cancelled
      await orderAPI.updateOrderStatus(orderId, "cancelled");
      await orderAPI.deleteOrder(orderId);

      // Then remove from local state
      setRecentOrders(recentOrders.filter((order) => order.id !== orderId));
      closeOrderModal();
    } catch (error) {
      console.error("Error deleting order:", error);
      setError("Failed to delete order: " + (error.message || "Unknown error"));
    } finally {
      setUpdatingStatus(false);
    }
  };

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
                    <select
                      className={`status-select status-${order.status}`}
                      value={order.status}
                      onChange={(e) =>
                        handleStatusChange(order.id, e.target.value)
                      }
                      disabled={
                        updatingStatus ||
                        order.status === "completed" ||
                        order.status === "cancelled"
                      }
                    >
                      <option value="pending">Pending</option>
                      <option value="completed">Completed</option>
                      <option value="cancelled">Cancelled</option>
                    </select>
                  </td>
                  <td>${order.total_amount.toFixed(2)}</td>
                  <td>
                    <button
                      className="edit-button"
                      onClick={() => openOrderModal(order)}
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showOrderModal && selectedOrder && (
        <div className="modal-overlay">
          <div className="modal-content">
            <h3>Edit Order #{selectedOrder.id}</h3>
            <OrderEditForm
              order={selectedOrder}
              onSubmit={updateOrder}
              onCancel={closeOrderModal}
              onDelete={handleDeleteOrder}
            />
          </div>
        </div>
      )}
    </div>
  );
};

const OrderEditForm = ({ order, onSubmit, onCancel, onDelete }) => {
  const [formData, setFormData] = useState({
    ...order,
    notes: order.notes || "",
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="order-edit-form">
      <div className="form-group">
        <label>Status:</label>
        <select
          name="status"
          value={formData.status}
          onChange={handleChange}
          className={`status-select status-${formData.status}`}
          disabled={
            formData.status === "completed" || formData.status === "cancelled"
          }
        >
          <option value="pending">Pending</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      <div className="form-group">
        <label>Notes:</label>
        <textarea
          name="notes"
          value={formData.notes}
          onChange={handleChange}
          rows="4"
          placeholder="Add notes about this order..."
        ></textarea>
      </div>

      <div className="form-actions">
        <button
          type="button"
          className="btn-delete"
          onClick={() => onDelete(order.id)}
        >
          Delete Order
        </button>
        <div className="right-actions">
          <button type="button" className="btn-secondary" onClick={onCancel}>
            Cancel
          </button>
          <button type="submit" className="btn-primary">
            Save Changes
          </button>
        </div>
      </div>
    </form>
  );
};

export default Dashboard;

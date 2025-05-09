import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { AuthContext } from "../../contexts/AuthContext";
import { useContext } from "react";
import { ingredientAPI } from "../../services/api";
import "../../styles/inventory.css";

const Inventory = () => {
  const { currentUser: user } = useContext(AuthContext);
  const [ingredients, setIngredients] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [sortField, setSortField] = useState("name");
  const [sortDirection, setSortDirection] = useState("asc");

  // Fetch ingredients on component mount
  useEffect(() => {
    const fetchIngredients = async () => {
      try {
        if (user && user.merchant_id) {
          setLoading(true);
          const response = await ingredientAPI.getIngredients(user.merchant_id);
          console.log("Ingredients response: ", response.data);
          setIngredients(response.data || []);
          setError(null);
        }
      } catch (err) {
        console.error("Error fetching ingredients:", err);
        setError("Failed to load ingredients. Please try again.");
      } finally {
        setLoading(false);
      }
    };

    fetchIngredients();
  }, [user]);

  // Handle ingredient deletion
  const handleDelete = async (id) => {
    if (window.confirm("Are you sure you want to delete this ingredient?")) {
      try {
        await ingredientAPI.deleteIngredient(user.merchant_id, id);
        setIngredients(
          ingredients.filter((ingredient) => ingredient.id !== id)
        );
      } catch (err) {
        console.error("Error deleting ingredient:", err);
        setError("Failed to delete ingredient. Please try again.");
      }
    }
  };

  // Filter ingredients based on search term
  const filteredIngredients = ingredients.filter((ingredient) =>
    ingredient.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Sort ingredients
  const sortedIngredients = [...filteredIngredients].sort((a, b) => {
    const aValue = a[sortField];
    const bValue = b[sortField];

    if (typeof aValue === "string") {
      return sortDirection === "asc"
        ? aValue.localeCompare(bValue)
        : bValue.localeCompare(aValue);
    } else {
      return sortDirection === "asc" ? aValue - bValue : bValue - aValue;
    }
  });

  // Toggle sort direction
  const toggleSortDirection = () => {
    setSortDirection(sortDirection === "asc" ? "desc" : "asc");
  };

  if (loading) return <div className="loading">Loading inventory...</div>;

  return (
    <div className="inventory-page">
      <h1>Ingredient Inventory</h1>

      {error && <div className="error-message">{error}</div>}

      <div className="inventory-controls">
        <input
          type="text"
          placeholder="Search ingredients..."
          className="search-box"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />

        <div className="sort-controls">
          <select
            className="sort-select"
            value={sortField}
            onChange={(e) => setSortField(e.target.value)}
          >
            <option value="name">Name</option>
            <option value="quantity">Quantity</option>
            <option value="unit">Unit</option>
          </select>

          <button className="sort-direction" onClick={toggleSortDirection}>
            {sortDirection === "asc" ? "↑" : "↓"}
          </button>
        </div>

        <Link to="/merchant/inventory/add" className="button primary">
          Add New Ingredient
        </Link>
      </div>

      {sortedIngredients.length === 0 ? (
        <div className="empty-state">
          <p>No ingredients found. Add some to manage your inventory.</p>
        </div>
      ) : (
        <table className="inventory-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Quantity</th>
              <th>Unit</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {sortedIngredients.map((ingredient) => (
              <tr key={ingredient.id}>
                <td>{ingredient.name}</td>
                <td>{ingredient.quantity}</td>
                <td>{ingredient.unit}</td>
                <td>
                  <span
                    className={`status-badge ${
                      ingredient.quantity <= ingredient.low_stock_threshold
                        ? "status-low"
                        : "status-ok"
                    }`}
                  >
                    {ingredient.quantity <= ingredient.low_stock_threshold
                      ? "Low Stock"
                      : "In Stock"}
                  </span>
                </td>
                <td>
                  <Link
                    to={`/merchant/inventory/${ingredient.id}`}
                    className="action-link"
                  >
                    Edit
                  </Link>
                  {" | "}
                  <button
                    onClick={() => handleDelete(ingredient.id)}
                    className="action-link"
                    style={{
                      background: "none",
                      border: "none",
                      cursor: "pointer",
                      color: "#d32f2f",
                    }}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default Inventory;

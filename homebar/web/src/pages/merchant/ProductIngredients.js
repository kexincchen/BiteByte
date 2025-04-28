import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
// import { useAuth } from '../../contexts/AuthContext';
import { AuthContext } from "../../contexts/AuthContext";
import { useContext } from "react";
import { productAPI, ingredientAPI, productIngredientAPI } from '../../services/api';
import '../../styles/inventory.css';

const ProductIngredients = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  // const { user } = useAuth();
  const { currentUser: user } = useContext(AuthContext);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [product, setProduct] = useState(null);
  const [ingredients, setIngredients] = useState([]);
  const [availableIngredients, setAvailableIngredients] = useState([]);
  const [newIngredient, setNewIngredient] = useState({
    ingredient_id: '',
    quantity: 0
  });

  // Fetch data on component mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch product details
        const productResponse = await productAPI.getProduct(id);
        setProduct(productResponse.data);
        
        // Fetch ingredients used in this product
        const ingredientsResponse = await productIngredientAPI.getProductIngredients(id);
        setIngredients(ingredientsResponse.data);
        
        // Fetch all merchant's ingredients for the dropdown
        const allIngredientsResponse = await ingredientAPI.getIngredients(user.merchant_id);
        setAvailableIngredients(allIngredientsResponse.data);
        
      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Failed to load data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id, user.merchant_id]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setNewIngredient({
      ...newIngredient,
      [name]: name === 'quantity' ? parseFloat(value) : parseInt(value, 10)
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      // Prepare data with the product ID
      const data = {
        product_id: parseInt(id, 10),
        ingredient_id: newIngredient.ingredient_id,
        quantity: newIngredient.quantity
      };
      
      await productIngredientAPI.addIngredientToProduct(id, data);
      
      // Refresh the ingredients list
      const response = await productIngredientAPI.getProductIngredients(id);
      setIngredients(response.data);
      
      // Reset the form
      setNewIngredient({
        ingredient_id: '',
        quantity: 0
      });
      
    } catch (err) {
      console.error('Error adding ingredient to product:', err);
      setError('Failed to add ingredient to product. Please try again.');
    }
  };

  const handleRemoveIngredient = async (ingredientId) => {
    if (window.confirm('Are you sure you want to remove this ingredient from the product?')) {
      try {
        await productIngredientAPI.removeIngredientFromProduct(id, ingredientId);
        
        // Update the UI by filtering out the removed ingredient
        setIngredients(ingredients.filter(ing => ing.ingredient_id !== ingredientId));
        
      } catch (err) {
        console.error('Error removing ingredient:', err);
        setError('Failed to remove ingredient. Please try again.');
      }
    }
  };

  // Filter out ingredients already added to the product
  const filteredAvailableIngredients = availableIngredients.filter(ai => 
    !ingredients.some(i => i.ingredient_id === ai.id)
  );

  if (loading) return <div className="loading">Loading...</div>;
  if (!product) return <div className="error-message">Product not found</div>;

  return (
    <div className="inventory-page">
      <h1>{product.name} - Ingredients</h1>
      
      {error && <div className="error-message">{error}</div>}
      
      <div className="product-details">
        <p><strong>Description:</strong> {product.description}</p>
        <p><strong>Price:</strong> ${product.price.toFixed(2)}</p>
      </div>
      
      <h2>Current Ingredients</h2>
      
      {ingredients.length === 0 ? (
        <div className="empty-state">
          <p>No ingredients added to this product yet.</p>
        </div>
      ) : (
        <table className="inventory-table">
          <thead>
            <tr>
              <th>Ingredient</th>
              <th>Quantity</th>
              <th>Unit</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {ingredients.map((ingredient) => {
              // Find the corresponding ingredient details from available ingredients
              const ingredientDetails = availableIngredients.find(i => i.id === ingredient.ingredient_id);
              return (
                <tr key={ingredient.ingredient_id}>
                  <td>{ingredientDetails ? ingredientDetails.name : 'Unknown'}</td>
                  <td>{ingredient.quantity}</td>
                  <td>{ingredientDetails ? ingredientDetails.unit : 'N/A'}</td>
                  <td>
                    <button 
                      onClick={() => handleRemoveIngredient(ingredient.ingredient_id)} 
                      className="action-link"
                      style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#d32f2f' }}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
      
      <h2>Add Ingredient</h2>
      
      {filteredAvailableIngredients.length === 0 ? (
        <div className="empty-state">
          <p>All available ingredients have been added to this product.</p>
        </div>
      ) : (
        <form className="ingredient-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="ingredient_id">Select Ingredient:</label>
            <select
              id="ingredient_id"
              name="ingredient_id"
              value={newIngredient.ingredient_id}
              onChange={handleChange}
              required
            >
              <option value="">-- Select an ingredient --</option>
              {filteredAvailableIngredients.map(ingredient => (
                <option key={ingredient.id} value={ingredient.id}>
                  {ingredient.name} ({ingredient.unit})
                </option>
              ))}
            </select>
          </div>
          
          <div className="form-group">
            <label htmlFor="quantity">Quantity:</label>
            <input
              type="number"
              id="quantity"
              name="quantity"
              min="0.01"
              step="0.01"
              value={newIngredient.quantity}
              onChange={handleChange}
              required
            />
            <small>
              How much of this ingredient is needed for one product?
            </small>
          </div>
          
          <div className="form-actions">
            <button type="button" className="button secondary" onClick={() => navigate(`/merchant/products/edit/${id}`)}>
              Back to Product
            </button>
            <button type="submit" className="button primary">
              Add Ingredient
            </button>
          </div>
        </form>
      )}
    </div>
  );
};

export default ProductIngredients; 
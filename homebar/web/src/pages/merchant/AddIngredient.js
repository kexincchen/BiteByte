import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from "../../contexts/AuthContext";
import { useContext } from "react";
import { ingredientAPI } from '../../services/api';
import '../../styles/inventory.css';

const AddIngredient = () => {
  const { currentUser: user } = useContext(AuthContext);
  const navigate = useNavigate();
  const [error, setError] = useState(null);
  const [ingredient, setIngredient] = useState({
    name: '',
    quantity: 0,
    unit: '',
    low_stock_threshold: 0
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setIngredient({
      ...ingredient,
      [name]: name === 'quantity' || name === 'low_stock_threshold' ? parseFloat(value) : value
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      await ingredientAPI.createIngredient(user.merchant_id, ingredient);
      navigate('/merchant/inventory');
    } catch (err) {
      console.error('Error adding ingredient:', err);
      setError('Failed to add ingredient. Please try again.');
    }
  };

  return (
    <div className="add-ingredient-page">
      <h1>Add New Ingredient</h1>
      
      {error && <div className="error-message">{error}</div>}
      
      <form className="ingredient-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="name">Ingredient Name:</label>
          <input
            type="text"
            id="name"
            name="name"
            value={ingredient.name}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="quantity">Initial Quantity:</label>
          <input
            type="number"
            id="quantity"
            name="quantity"
            min="0"
            step="0.01"
            value={ingredient.quantity}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="unit">Unit of Measurement:</label>
          <input
            type="text"
            id="unit"
            name="unit"
            placeholder="e.g., ml, g, pieces"
            value={ingredient.unit}
            onChange={handleChange}
            required
          />
          <small>Enter the unit of measurement for this ingredient (e.g., ml, grams, pieces).</small>
        </div>
        
        <div className="form-group">
          <label htmlFor="low_stock_threshold">Low Stock Threshold:</label>
          <input
            type="number"
            id="low_stock_threshold"
            name="low_stock_threshold"
            min="0"
            step="0.01"
            value={ingredient.low_stock_threshold}
            onChange={handleChange}
            required
          />
          <small>You'll be alerted when stock falls below this level.</small>
        </div>
        
        <div className="form-actions">
          <button type="button" className="button secondary" onClick={() => navigate('/merchant/inventory')}>
            Cancel
          </button>
          <button type="submit" className="button primary">
            Add Ingredient
          </button>
        </div>
      </form>
    </div>
  );
};

export default AddIngredient; 

import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { AuthContext } from "../../contexts/AuthContext";
import { useContext } from "react";
import { ingredientAPI } from '../../services/api';
import '../../styles/inventory.css';

const EditIngredient = () => {
  const { id } = useParams();
  const { currentUser: user } = useContext(AuthContext);
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [ingredient, setIngredient] = useState({
    name: '',
    quantity: 0,
    unit: '',
    low_stock_threshold: 0
  });

  // Fetch ingredient data on component mount
  useEffect(() => {
    const fetchIngredient = async () => {
      try {
        setLoading(true);
        const response = await ingredientAPI.getIngredients(user.merchant_id);
        const foundIngredient = response.data.find(item => item.id === parseInt(id));
        
        if (foundIngredient) {
          setIngredient(foundIngredient);
        } else {
          setError('Ingredient not found');
        }
      } catch (err) {
        console.error('Error fetching ingredient:', err);
        setError('Failed to load ingredient data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchIngredient();
  }, [id, user.merchant_id]);

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
      await ingredientAPI.updateIngredient(user.merchant_id, ingredient.id, ingredient);
      navigate('/merchant/inventory');
    } catch (err) {
      console.error('Error updating ingredient:', err);
      setError('Failed to update ingredient. Please try again.');
    }
  };

  if (loading) return <div className="loading">Loading ingredient data...</div>;

  return (
    <div className="add-ingredient-page">
      <h1>Edit Ingredient</h1>
      
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
          <label htmlFor="quantity">Quantity:</label>
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
            value={ingredient.unit}
            onChange={handleChange}
            required
          />
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
            Save Changes
          </button>
        </div>
      </form>
    </div>
  );
};

export default EditIngredient; 

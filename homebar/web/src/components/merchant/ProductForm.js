import React, { useState, useEffect, useContext } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { productAPI } from '../../services/api';
import { AuthContext } from '../../contexts/AuthContext';

const ProductForm = ({ isEditing = false }) => {
  const { currentUser } = useContext(AuthContext);
  const navigate = useNavigate();
  const { id } = useParams();

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    price: '',
    category: '',
    image_url: '',
    is_available: true
  });
  
  const [loading, setLoading] = useState(isEditing);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [categories, setCategories] = useState([
    'Cocktails', 'Spirits', 'Beer', 'Wine', 'Non-Alcoholic', 'Snacks'
  ]);

  useEffect(() => {
    const fetchProductCategories = async () => {
      try {
        // In a real app, you would fetch categories from the server
        // For now, we'll use the hardcoded categories
      } catch (error) {
        console.error('Error fetching categories:', error);
      }
    };

    const fetchProductDetails = async () => {
      if (isEditing && id) {
        try {
          const response = await productAPI.getProduct(id);
          setFormData(response.data);
          setLoading(false);
        } catch (error) {
          console.error('Error fetching product details:', error);
          setError('Failed to load product details');
          setLoading(false);
        }
      }
    };

    fetchProductCategories();
    if (isEditing) {
      fetchProductDetails();
    }
  }, [isEditing, id]);

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!currentUser || currentUser.role !== 'merchant') {
      setError('You must be logged in as a merchant to perform this action');
      return;
    }

    setSubmitting(true);
    setError('');

    try {
      // Format the data for submission
      const productData = {
        ...formData,
        price: parseFloat(formData.price),
        merchant_id: currentUser.merchant_id
      };

      if (isEditing) {
        await productAPI.updateProduct(id, productData);
      } else {
        await productAPI.createProduct(productData);
      }

      navigate('/merchant/products');
    } catch (error) {
      console.error('Error saving product:', error);
      setError('Failed to save product. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div>Loading product details...</div>;

  return (
    <div className="product-form-container">
      <h1>{isEditing ? 'Edit Product' : 'Add New Product'}</h1>
      
      {error && <div className="error">{error}</div>}
      
      <form onSubmit={handleSubmit} className="product-form">
        <div className="form-group">
          <label htmlFor="name">Product Name</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows="4"
            required
          />
        </div>
        
        <div className="form-row">
          <div className="form-group">
            <label htmlFor="price">Price ($)</label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price}
              onChange={handleChange}
              step="0.01"
              min="0"
              required
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="category">Category</label>
            <select
              id="category"
              name="category"
              value={formData.category}
              onChange={handleChange}
              required
            >
              <option value="">Select a category</option>
              {categories.map(category => (
                <option key={category} value={category}>{category}</option>
              ))}
            </select>
          </div>
        </div>
        
        <div className="form-group">
          <label htmlFor="image_url">Image URL</label>
          <input
            type="text"
            id="image_url"
            name="image_url"
            value={formData.image_url}
            onChange={handleChange}
            placeholder="https://example.com/image.jpg"
          />
          {formData.image_url && (
            <div className="image-preview">
              <img src={formData.image_url} alt="Product preview" />
            </div>
          )}
        </div>
        
        <div className="form-group checkbox-group">
          <input
            type="checkbox"
            id="is_available"
            name="is_available"
            checked={formData.is_available}
            onChange={handleChange}
          />
          <label htmlFor="is_available">Available for purchase</label>
        </div>
        
        <div className="form-actions">
          <button type="button" onClick={() => navigate('/merchant/products')} className="cancel-button">
            Cancel
          </button>
          <button type="submit" disabled={submitting} className="submit-button">
            {submitting ? 'Saving...' : isEditing ? 'Update Product' : 'Add Product'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ProductForm; 
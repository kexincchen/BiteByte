import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { productAPI } from '../services/api';

const Products = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [category, setCategory] = useState('');
  const [categories, setCategories] = useState([]);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await productAPI.getProducts();
        const fetchedProducts = response.data;
        
        setProducts(fetchedProducts);
        
        // Extract unique categories
        const uniqueCategories = [...new Set(fetchedProducts.map(p => p.category))];
        setCategories(uniqueCategories);
        
        setLoading(false);
      } catch (error) {
        console.error('Error fetching products:', error);
        setError('Failed to fetch products. Please try again later.');
        setLoading(false);
      }
    };
    
    fetchProducts();
  }, []);

  const filteredProducts = category 
    ? products.filter(p => p.category === category) 
    : products;

  if (loading) return <div>Loading products...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="products-page">
      <h1>Drinks Menu</h1>
      
      <div className="category-filter">
        <button 
          className={category === '' ? 'active' : ''} 
          onClick={() => setCategory('')}
        >
          All
        </button>
        {categories.map(cat => (
          <button 
            key={cat} 
            className={category === cat ? 'active' : ''} 
            onClick={() => setCategory(cat)}
          >
            {cat}
          </button>
        ))}
      </div>
      
      <div className="products-grid">
        {filteredProducts.map(product => (
          <div key={product.id} className="product-card">
            <div className="product-image">
              <img
              src={productAPI.imageUrl(product.id)}
              alt={product.name}
              onError={(e) => {e.target.src = '/placeholder.png'}}
              />
            </div>
            <div className="product-details">
              <h3>{product.name}</h3>
              <p className="product-description">{product.description}</p>
              <p className="product-price">${product.price.toFixed(2)}</p>
              <Link to={`/products/${product.id}`} className="view-button">
                View Details
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Products;

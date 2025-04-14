import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';

const Products = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [category, setCategory] = useState('');
  const [categories, setCategories] = useState([]);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        // In a real app, this would be a real API call
        // const response = await axios.get('/api/products');
        
        // For demo, simulating API response
        const mockProducts = [
          {
            id: 1,
            name: 'Mojito',
            description: 'Classic cocktail with rum, mint, and lime',
            price: 8.99,
            category: 'Cocktails',
            imageUrl: 'https://example.com/mojito.jpg',
            isAvailable: true
          },
          {
            id: 2,
            name: 'Old Fashioned',
            description: 'Whiskey cocktail with sugar and bitters',
            price: 9.99,
            category: 'Cocktails',
            imageUrl: 'https://example.com/old-fashioned.jpg',
            isAvailable: true
          },
          {
            id: 3,
            name: 'Margarita',
            description: 'Tequila cocktail with lime and salt',
            price: 7.99,
            category: 'Cocktails',
            imageUrl: 'https://example.com/margarita.jpg',
            isAvailable: true
          }
        ];
        
        setProducts(mockProducts);
        
        // Extract unique categories
        const uniqueCategories = [...new Set(mockProducts.map(p => p.category))];
        setCategories(uniqueCategories);
        
        setLoading(false);
      } catch (error) {
        setError('Failed to fetch products');
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
              <img src={product.imageUrl} alt={product.name} />
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
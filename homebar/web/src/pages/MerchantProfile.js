import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { productAPI } from '../services/api';

const MerchantProfile = () => {
  const [merchant, setMerchant] = useState(null);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { username } = useParams();

  useEffect(() => {
    const fetchMerchantData = async () => {
      try {
        // Fetch merchant profile
        const merchantResponse = await productAPI.getMerchantByUsername(username);
        setMerchant(merchantResponse.data);
        
        // Fetch merchant's products
        const productsResponse = await productAPI.getProductsByMerchant(merchantResponse.data.id);
        setProducts(productsResponse.data);
        
        setLoading(false);
      } catch (error) {
        console.error('Error fetching merchant data:', error);
        setError('Could not load merchant profile');
        setLoading(false);
      }
    };
    
    fetchMerchantData();
  }, [username]);

  if (loading) return <div>Loading merchant profile...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!merchant) return <div className="error">Merchant not found</div>;

  return (
    <div className="merchant-profile-page">
      <div className="merchant-header">
        <h1>{merchant.business_name}</h1>
        <div className="merchant-info">
          <p className="merchant-description">{merchant.description}</p>
          <div className="merchant-details">
            <p><strong>Address:</strong> {merchant.address}</p>
            <p><strong>Phone:</strong> {merchant.phone}</p>
          </div>
        </div>
      </div>
      
      <h2>Menu</h2>
      {products && products.length === 0 ? (
        <p>This merchant doesn't have any products yet.</p>
      ) : (
        <div className="products-grid">
          {products && products.map(product => (
            <div key={product.id} className="product-card">
              <div className="product-image">
                <img src={product.image_url} alt={product.name} />
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
      )}
    </div>
  );
};

export default MerchantProfile; 
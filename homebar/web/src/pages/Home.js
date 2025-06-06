import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { merchantAPI } from '../services/api';

const Home = () => {
  const [merchants, setMerchants] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMerchants = async () => {
      try {
        const response = await merchantAPI.getMerchants();
        // Ensure we have an array, even if response.data is null or undefined
        setMerchants(response.data || []);
        setLoading(false);
      } catch (error) {
        console.error('Error fetching merchants:', error);
        setLoading(false);
      }
    };

    fetchMerchants();
  }, []);

  return (
    <div className="home-page">
      <div className="hero-section">
        <h1>Welcome to Home Bar</h1>
        <p>Your one-stop solution for ordering drinks and cocktails from home</p>
        <Link to="/products" className="cta-button">
          Browse Menu
        </Link>
      </div>
      
      <div className="merchants-section">
        <h2>Featured Merchants</h2>
        {loading ? (
          <p>Loading merchants...</p>
        ) : (
          <div className="merchants-grid">
            {merchants.map(merchant => (
              <div key={merchant.id} className="merchant-card">
                <h3>{merchant.business_name}</h3>
                <p>{merchant.description.substring(0, 100)}...</p>
                <Link to={`/${merchant.username}`} className="view-merchant-button">
                  View Menu
                </Link>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Home; 
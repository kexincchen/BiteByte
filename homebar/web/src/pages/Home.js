import React from 'react';
import { Link } from 'react-router-dom';

const Home = () => {
  return (
    <div className="home-page">
      <div className="hero-section">
        <h1>Welcome to Home Bar</h1>
        <p>Your one-stop solution for ordering drinks and cocktails from home</p>
        <Link to="/products" className="cta-button">
          Browse Menu
        </Link>
      </div>
      
      <div className="features-section">
        <h2>Our Features</h2>
        <div className="features-grid">
          <div className="feature-card">
            <h3>Wide Selection</h3>
            <p>Choose from a variety of cocktails and drinks</p>
          </div>
          <div className="feature-card">
            <h3>Fast Delivery</h3>
            <p>Get your orders delivered right to your doorstep</p>
          </div>
          <div className="feature-card">
            <h3>Quality Ingredients</h3>
            <p>We use only the finest ingredients for our drinks</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home; 
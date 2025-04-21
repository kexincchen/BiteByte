import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import Inventory from './pages/merchant/Inventory';
import AddIngredient from './pages/merchant/AddIngredient';
import ProductIngredients from './pages/merchant/ProductIngredients';
import './styles/inventory.css';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
); 
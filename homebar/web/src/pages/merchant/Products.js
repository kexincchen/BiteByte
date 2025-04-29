import React, { useState, useEffect, useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { productAPI } from '../../services/api';
import { AuthContext } from '../../contexts/AuthContext';

const Products = () => {
  const { currentUser } = useContext(AuthContext);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [categories, setCategories] = useState([]);
  const navigate = useNavigate();

  const fetchProducts = async () => {
    if (!currentUser || currentUser.role !== 'merchant') {
      setError('You must be logged in as a merchant to view this page');
      setLoading(false);
      return;
    }

    try {
      const response = await productAPI.getProductsByMerchant(currentUser.merchant_id);
      setProducts(response.data);
      
      // Extract unique categories
      const uniqueCategories = [...new Set(response.data.map(p => p.category))];
      setCategories(uniqueCategories);
      
      setLoading(false);
    } catch (error) {
      console.error('Error fetching products:', error);
      setError('Failed to load products');
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProducts();
  }, [currentUser]);

  const handleToggleAvailability = async (product) => {
    try {
      const updatedProduct = { ...product, is_available: !product.is_available };
      await productAPI.updateProduct(product.id, updatedProduct);
      
      // Update the products state to reflect the change
      setProducts(products.map(p => 
        p.id === product.id ? { ...p, is_available: !p.is_available } : p
      ));
    } catch (error) {
      console.error('Error updating product availability:', error);
      alert('Failed to update product availability. Please try again.');
    }
  };

  // Filter products by search term and category
  const filteredProducts = products
    .filter(p => p.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                p.description.toLowerCase().includes(searchTerm.toLowerCase()))
    .filter(p => categoryFilter ? p.category === categoryFilter : true);

  if (loading) return <div>Loading products...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="merchant-products-page">
      <div className="page-header">
        <h1>Manage Menu</h1>
        <Link to="/merchant/products/new" className="add-product-button">
          Add New Product
        </Link>
      </div>
      
      <div className="filters-section">
        <div className="search-box">
          <input
            type="text"
            placeholder="Search products..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        
        <div className="category-filter">
          <select 
            value={categoryFilter} 
            onChange={(e) => setCategoryFilter(e.target.value)}
          >
            <option value="">All Categories</option>
            {categories.map(category => (
              <option key={category} value={category}>{category}</option>
            ))}
          </select>
        </div>
      </div>
      
      {filteredProducts.length === 0 ? (
        <div className="no-products">
          <p>No products found. {categoryFilter && 'Try changing the category filter or '} 
          <Link to="/merchant/products/new">add a new product</Link>.</p>
        </div>
      ) : (
        <div className="products-table-container">
          <table className="products-table">
            <thead>
              <tr>
                <th>Image</th>
                <th>Name</th>
                <th>Category</th>
                <th>Price</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredProducts.map(product => (
                <tr key={product.id}>
                  <td className="product-image-cell">
                    <img
                        src={`http://localhost:8080/api/products/${product.id}/image`}
                        alt={product.name}
                        className="product-thumbnail"
                        onError={(e) => {
                          e.target.src = '/placeholder.png'
                        }}
                    />
                  </td>
                  <td>{product.name}</td>
                  <td>{product.category}</td>
                  <td>${product.price.toFixed(2)}</td>
                  <td>
                    <span
                        className={`status-badge ${product.is_available ? 'status-available' : 'status-unavailable'}`}
                      onClick={() => handleToggleAvailability(product)}
                    >
                      {product.is_available ? 'Available' : 'Unavailable'}
                    </span>
                  </td>
                  <td className="actions-cell">
                    <button 
                      className="edit-button"
                      onClick={() => navigate(`/merchant/products/edit/${product.id}`)}
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default Products;

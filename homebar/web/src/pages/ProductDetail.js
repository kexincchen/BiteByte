import React, { useState, useEffect, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { productAPI } from "../services/api";
import { CartContext } from "../contexts/CartContext";

const ProductDetail = () => {
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [quantity, setQuantity] = useState(1);
  const [isAvailable, setIsAvailable] = useState(false);
  const [checkingAvailability, setCheckingAvailability] = useState(false);

  const { id } = useParams();
  const navigate = useNavigate();
  const { addToCart } = useContext(CartContext);

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        const response = await productAPI.getProduct(id);
        setProduct(response.data);
        setLoading(false);

        // Check real-time availability after getting product data
        if (response.data) {
          checkProductAvailability(response.data.id);
        }
      } catch (error) {
        console.error("Error fetching product:", error);
        setError("Failed to load product details");
        setLoading(false);
      }
    };

    fetchProduct();
  }, [id]);

  const checkProductAvailability = async (productId) => {
    setCheckingAvailability(true);
    try {
      const response = await productAPI.checkAvailability([productId]);
      setIsAvailable(response.data.availability[productId]);
    } catch (error) {
      console.error("Error checking product availability:", error);
      // Fall back to the static availability flag if API check fails
      setIsAvailable(product?.is_available || false);
    } finally {
      setCheckingAvailability(false);
    }
  };

  const handleAddToCart = () => {
    addToCart(product, quantity);
    navigate("/cart");
  };

  if (loading) return <div>Loading product details...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!product) return <div className="error">Product not found</div>;

  // Determine final availability status, with priority to real-time check
  const productAvailable = checkingAvailability ? false : isAvailable;

  return (
    <div className="product-detail-page">
      <div className="product-detail-container">
        <div className="product-image-container">
          <img
              src={productAPI.imageUrl(product.id)}
              alt={product.name}
              onError={(e) => {
                e.target.src = '/placeholder.png'
              }}
          />
        </div>

        <div className="product-info">
          <h1>{product.name}</h1>
          <p className="product-description">{product.description}</p>
          <p className="product-price">${product.price.toFixed(2)}</p>

          {product.ingredients && (
            <div className="product-ingredients">
              <h3>Ingredients</h3>
              <ul>
                {product.ingredients.map((ingredient) => (
                  <li key={ingredient.id}>
                    {ingredient.name} ({ingredient.quantity} {ingredient.unit})
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className="product-actions">
            <div className="quantity-selector">
              <button
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
                disabled={quantity <= 1}
              >
                -
              </button>
              <span>{quantity}</span>
              <button onClick={() => setQuantity(quantity + 1)}>+</button>
            </div>

            <button
              className="add-to-cart-button"
              onClick={handleAddToCart}
              disabled={!productAvailable || checkingAvailability}
            >
              {checkingAvailability
                ? "Checking availability..."
                : productAvailable
                ? "Add to Cart"
                : "Out of Stock"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProductDetail;

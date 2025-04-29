import React, { useState, useEffect } from "react";
import { checkProductsAvailability } from "../services/api";

const Checkout = () => {
  const [cart, setCart] = useState([]);
  const [isProcessing, setIsProcessing] = useState(false);
  const [productAvailability, setProductAvailability] = useState({});
  const [isLoadingAvailability, setIsLoadingAvailability] = useState(false);

  const checkAvailability = async () => {
    if (!cart || cart.length === 0) return;

    setIsLoadingAvailability(true);
    try {
      const productIds = cart.map((item) => item.product.id);
      const availability = await checkProductsAvailability(productIds);
      setProductAvailability(availability);
    } catch (error) {
      console.error("Failed to check product availability", error);
    } finally {
      setIsLoadingAvailability(false);
    }
  };

  useEffect(() => {
    checkAvailability();
  }, [cart]);

  const handleCheckout = () => {
    setIsProcessing(true);
    // Add checkout logic here
  };

  return (
    <div>
      <h1>Checkout Page</h1>
      <p>This is a placeholder for the checkout page.</p>
      {cart.map((item) => (
        <div key={item.product.id} className="cart-item">
          <div className="cart-item-details">
            <h3>{item.product.name}</h3>
            <p>${item.product.price.toFixed(2)}</p>
            {productAvailability[item.product.id] === false && (
              <span className="sold-out-badge">Sold Out</span>
            )}
          </div>
          {/* ... rest of cart item display */}
        </div>
      ))}
      <button
        onClick={handleCheckout}
        disabled={
          isProcessing ||
          cart.length === 0 ||
          cart.some((item) => productAvailability[item.product.id] === false)
        }
        className="checkout-button"
      >
        {isProcessing ? "Processing..." : "Checkout"}
      </button>
    </div>
  );
};

export default Checkout;

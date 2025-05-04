import React, { useContext, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CartContext } from "../contexts/CartContext";
import { AuthContext } from "../contexts/AuthContext";
import {orderAPI, productAPI} from "../services/api";

const Cart = () => {
  const { cartItems, cartTotal, removeFromCart, updateQuantity, clearCart } =
    useContext(CartContext);
  const { currentUser } = useContext(AuthContext);
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleCheckout = async () => {
    if (!currentUser) {
      navigate("/login", { state: { from: "/cart" } });
      return;
    }

    setIsProcessing(true);
    setError("");

    try {
      // Format the order data
      const orderData = {
        customer_id: parseInt(currentUser.id), // Ensure it's a number
        // merchant_id:cartItems[0].merchant_id, // Assuming all items are from the same
        merchant_id: 1, // Use a default merchant_id for demo purposes
        items: cartItems.map((item) => ({
          product_id: parseInt(item.id), // Ensure it's a number
          quantity: item.quantity,
          price: item.price,
        })),
        notes: "",
      };

      console.log("Sending order data:", orderData); // Debug log

      // Create the order
      const response = await orderAPI.createOrder(orderData);

      // Clear the cart after successful order
      clearCart();

      // Navigate to the order confirmation
      navigate(`/orders/${response.data.id}`, {
        state: { order: response.data },
      });
    } catch (error) {
      console.error("Checkout error:", error);
      setError("Failed to process your order. Please try again.");
    } finally {
      setIsProcessing(false);
    }
  };

  if (cartItems.length === 0) {
    return (
      <div className="cart-page">
        <h1>Your Cart</h1>
        <p>Your cart is empty.</p>
        <Link to="/products" className="button">
          Browse Products
        </Link>
      </div>
    );
  }

  return (
    <div className="cart-page">
      <h1>Your Cart</h1>

      {error && <div className="error">{error}</div>}

      <div className="cart-items">
        {cartItems.map((item) => (
          <div key={item.id} className="cart-item">
            <div className="item-image">
              <img
                  src={productAPI.imageUrl(item.id)}
                  alt={item.name}
                  onError={(e) => {
                    e.target.src = '/placeholder.png'
                  }}
              />
            </div>
            <div className="item-details">
              <h3>{item.name}</h3>
              <p>${item.price.toFixed(2)}</p>
            </div>
            <div className="item-quantity">
              <button
                onClick={() =>
                  updateQuantity(item.id, Math.max(1, item.quantity - 1))
                }
                disabled={item.quantity <= 1}
              >
                -
              </button>
              <span>{item.quantity}</span>
              <button
                onClick={() => updateQuantity(item.id, item.quantity + 1)}
              >
                +
              </button>
            </div>
            <div className="item-total">
              ${(item.price * item.quantity).toFixed(2)}
            </div>
            <button
              className="remove-button"
              onClick={() => removeFromCart(item.id)}
            >
              Remove
            </button>
          </div>
        ))}
      </div>

      <div className="cart-summary">
        <div className="cart-total">
          <span>Total:</span>
          <span>${cartTotal.toFixed(2)}</span>
        </div>

        <div className="cart-actions">
          <button className="clear-button" onClick={clearCart}>
            Clear Cart
          </button>
          <button
            className="checkout-button"
            onClick={handleCheckout}
            disabled={isProcessing}
          >
            {isProcessing ? "Processing..." : "Proceed to Checkout"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Cart;

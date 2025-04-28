import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Navbar from "./components/Navbar";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Products from "./pages/Products";
import ProductDetail from "./pages/ProductDetail";
import Cart from "./pages/Cart";
import Checkout from "./pages/Checkout";
import OrderHistory from "./pages/OrderHistory";
import OrderDetail from "./pages/OrderDetail";
import Profile from "./pages/Profile";
import MerchantProfile from "./pages/MerchantProfile";
import MerchantDashboard from "./pages/merchant/Dashboard";
import MerchantProducts from "./pages/merchant/Products";
import MerchantAddProduct from "./pages/merchant/AddProduct";
import MerchantEditProduct from "./pages/merchant/EditProduct";
import MerchantOrders from "./pages/merchant/Orders";
import MerchantInventory from "./pages/merchant/Inventory";
import AddIngredient from "./pages/merchant/AddIngredient";
import EditIngredient from "./pages/merchant/EditIngredient";
import ProductIngredients from "./pages/merchant/ProductIngredients";
import { AuthProvider } from "./contexts/AuthContext";
import { CartProvider } from "./contexts/CartContext";
import "./App.css";

function App() {
  return (
    <Router>
      <AuthProvider>
        <CartProvider>
          <div className="App">
            <Navbar />
            <div className="container">
              <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route path="/products" element={<Products />} />
                <Route path="/products/:id" element={<ProductDetail />} />
                <Route path="/cart" element={<Cart />} />
                <Route path="/checkout" element={<Checkout />} />
                <Route path="/orders" element={<OrderHistory />} />
                <Route path="/orders/:id" element={<OrderDetail />} />
                <Route path="/profile" element={<Profile />} />

                {/* New route for merchant profiles */}
                <Route path="/:username" element={<MerchantProfile />} />

                {/* Merchant Routes */}
                <Route path="/merchant" element={<MerchantDashboard />} />
                <Route
                  path="/merchant/products"
                  element={<MerchantProducts />}
                />
                <Route
                  path="/merchant/products/new"
                  element={<MerchantAddProduct />}
                />
                <Route
                  path="/merchant/products/edit/:id"
                  element={<MerchantEditProduct />}
                />
                <Route path="/merchant/orders" element={<MerchantOrders />} />
                <Route
                  path="/merchant/inventory"
                  element={<MerchantInventory />}
                />
                <Route
                  path="/merchant/inventory/add"
                  element={<AddIngredient />}
                />
                <Route
                  path="/merchant/inventory/:id"
                  element={<EditIngredient />}
                />
                <Route
                  path="/merchant/products/:id/ingredients"
                  element={<ProductIngredients />}
                />
              </Routes>
            </div>
          </div>
        </CartProvider>
      </AuthProvider>
    </Router>
  );
}

export default App;

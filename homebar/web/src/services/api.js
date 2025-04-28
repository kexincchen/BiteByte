import axios from "axios";

// Base API URL - in a real app, you'd use environment variables
const API_URL = "http://localhost:8080/api";

// Create an axios instance with default config
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Add a request interceptor to attach the auth token to every request
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  register: (userData) => {
    return apiClient.post("/auth/register", userData);
  },
  login: (email, password) => {
    return apiClient.post("/auth/login", { email, password });
  },
};

// Product API
export const productAPI = {
  getProducts: () => {
    return apiClient.get("/products");
  },
  getProduct: (id) => {
    return apiClient.get(`/products/${id}`);
  },
  createProduct: (productData) => {
    return apiClient.post("/products", productData);
  },
  updateProduct: (id, productData) => {
    return apiClient.put(`/products/${id}`, productData);
  },
  deleteProduct: (id) => {
    return apiClient.delete(`/products/${id}`);
  },
  getMerchantByUsername: (username) => {
    return apiClient.get(`/merchants/username/${username}`);
  },
  getProductsByMerchant: (merchantId) => {
    return apiClient.get(`/products/merchant/${merchantId}`);
  },
};

// Order API
export const orderAPI = {
  createOrder: (orderData) => {
    return apiClient.post("/orders", orderData);
  },
  getOrders: () => {
    return apiClient.get("/orders");
  },
  getOrder: (id) => {
    return apiClient.get(`/orders/${id}`);
  },
  updateOrderStatus: (id, status) => {
    return apiClient.put(`/orders/${id}/status`, { status });
  },
  updateOrder: (id, orderData) => {
    return apiClient.put(`/orders/${id}`, orderData);
  },
  getOrdersByMerchant: (merchantId) => {
    return apiClient.get(`/orders?merchant=${merchantId}`);
  },
  getOrdersByCustomer: (customerId) => {
    return apiClient.get(`/orders?customer=${customerId}`);
  },
};

// User/Profile API
export const userAPI = {
  getProfile: () => {
    return apiClient.get("/users/me");
  },
  updateProfile: (userData) => {
    return apiClient.put("/users/me", userData);
  },
};

// Cart API (if needed on server side)
export const cartAPI = {
  checkout: (cartData) => {
    return apiClient.post("/cart/checkout", cartData);
  },
};

// Add a dedicated Merchant API object if needed
export const merchantAPI = {
  getMerchants: () => {
    return apiClient.get("/merchants");
  },
  getMerchant: (id) => {
    return apiClient.get(`/merchants/${id}`);
  },
  getMerchantByUsername: (username) => {
    return apiClient.get(`/merchants/username/${username}`);
  },
};

// Add this utility function
export const fetchWithAuth = async (url, options = {}) => {
  const token = localStorage.getItem("token");

  const defaultOptions = {
    headers: {
      "Content-Type": "application/json",
      Authorization: token ? `Bearer ${token}` : "",
    },
  };

  const mergedOptions = {
    ...defaultOptions,
    ...options,
    headers: {
      ...defaultOptions.headers,
      ...(options.headers || {}),
    },
  };

  try {
    const response = await fetch(url, mergedOptions);

    // Check if the response is JSON
    const contentType = response.headers.get("content-type");
    const isJson = contentType && contentType.includes("application/json");

    // Parse the response
    const data = isJson ? await response.json() : await response.text();

    // Check if the response is successful
    if (!response.ok) {
      const error = new Error(
        isJson && data.error
          ? data.error
          : `HTTP error! Status: ${response.status}`
      );
      error.status = response.status;
      error.data = data;
      throw error;
    }

    return data;
  } catch (error) {
    console.error(`API error for ${url}:`, error);
    throw error;
  }
};

// Ingredient API methods
export const ingredientAPI = {
  getIngredients: (merchantId) => {
    return apiClient.get(`/merchants/${merchantId}/inventory`);
  },

  getInventorySummary: (merchantId) => {
    return apiClient.get(`/merchants/${merchantId}/inventory/summary`);
  },

  createIngredient: (merchantId, ingredientData) => {
    return apiClient.post(`/merchants/${merchantId}/inventory`, ingredientData);
  },

  updateIngredient: (merchantId, ingredientId, updateData) => {
    return apiClient.put(
      `/merchants/${merchantId}/inventory/${ingredientId}`,
      updateData
    );
  },

  deleteIngredient: (merchantId, ingredientId) => {
    return apiClient.delete(
      `/merchants/${merchantId}/inventory/${ingredientId}`
    );
  },
};

// Product Ingredient API methods
export const productIngredientAPI = {
  getProductIngredients: (productId) => {
    return apiClient.get(`/products/${productId}/ingredients`);
  },

  addIngredientToProduct: (productId, ingredientData) => {
    return apiClient.post(`/products/${productId}/ingredients`, ingredientData);
  },

  removeIngredientFromProduct: (productId, ingredientId) => {
    return apiClient.delete(
      `/products/${productId}/ingredients/${ingredientId}`
    );
  },

  updateProductIngredient: (productId, ingredientId, quantity) => {
    return apiClient.put(`/products/${productId}/ingredients/${ingredientId}`, {
      quantity,
    });
  },
};

export default {
  auth: authAPI,
  products: productAPI,
  orders: orderAPI,
  user: userAPI,
  cart: cartAPI,
  merchants: merchantAPI,
  ingredients: ingredientAPI,
  productIngredients: productIngredientAPI,
};

import axios from 'axios';

// Base API URL - in a real app, you'd use environment variables
const API_URL = 'http://localhost:8080/api';

// Create an axios instance with default config
const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add a request interceptor to attach the auth token to every request
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
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
    return apiClient.post('/auth/register', userData);
  },
  login: (email, password) => {
    return apiClient.post('/auth/login', { email, password });
  },
};

// Product API
export const productAPI = {
  getProducts: () => {
    return apiClient.get('/products');
  },
  getProduct: (id) => {
    return apiClient.get(`/products/${id}`);
  },
  createProduct: (productData) => {
    return apiClient.post('/products', productData);
  },
  updateProduct: (id, productData) => {
    return apiClient.put(`/products/${id}`, productData);
  },
  deleteProduct: (id) => {
    return apiClient.delete(`/products/${id}`);
  },
};

// Order API
export const orderAPI = {
  createOrder: (orderData) => {
    return apiClient.post('/orders', orderData);
  },
  getOrders: () => {
    return apiClient.get('/orders');
  },
  getOrder: (id) => {
    return apiClient.get(`/orders/${id}`);
  },
  updateOrderStatus: (id, status) => {
    return apiClient.patch(`/orders/${id}/status`, { status });
  },
};

// User/Profile API
export const userAPI = {
  getProfile: () => {
    return apiClient.get('/users/me');
  },
  updateProfile: (userData) => {
    return apiClient.put('/users/me', userData);
  },
};

// Cart API (if needed on server side)
export const cartAPI = {
  checkout: (cartData) => {
    return apiClient.post('/cart/checkout', cartData);
  },
};

export default {
  auth: authAPI,
  products: productAPI,
  orders: orderAPI,
  user: userAPI,
  cart: cartAPI,
}; 
import axios from "axios";

const HOSTS = (
  process.env.REACT_APP_SERVER_URLS ||
  "http://localhost:9001,http://localhost:9002,http://localhost:9003"
)
  .split(",")
  .map((h) => h.trim())
  .filter(Boolean);

let currIdx = 0;

function setBase(i) {
  currIdx = (i + HOSTS.length) % HOSTS.length;
  apiClient.defaults.baseURL = `${HOSTS[currIdx]}/api`;
}
function rotateBase() {
  setBase(currIdx + 1);
}
function getBaseURL() {
  return apiClient.defaults.baseURL;
}

const apiClient = axios.create();
setBase(0);

apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    if (!err.response) {
      const tried = currIdx;
      rotateBase();
      if (currIdx !== tried) {
        err.config.baseURL = getBaseURL();
        return apiClient(err.config);
      }
    }

    if (err.response && err.response.status === 307) {
      const loc = err.response.headers.location || "";
      try {
        const origin = new URL(loc).origin;
        const idx = HOSTS.indexOf(origin);
        if (idx !== -1) {
          setBase(idx);
        } else {
          apiClient.defaults.baseURL = origin + "/api";
        }
        err.config.baseURL = getBaseURL();
        return apiClient(err.config);
      } catch {}
    }

    return Promise.reject(err);
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
  checkAvailability: (productIds) => {
    return apiClient.post("/products/availability", {
      product_ids: productIds,
    });
  },
  imageUrl: (id, bust = true) =>
    `${getBaseURL()}/products/${id}/image${bust ? `?ts=${Date.now()}` : ""}`,
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
  deleteOrder: (id) => {
    return apiClient.delete(`/orders/${id}`);
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
    console.log("Getting product ingredients for product: ", productId);
    return apiClient.get(`/products/${productId}/ingredients`);
  },

  addIngredientToProduct: (productId, ingredientData) => {
    console.log("Adding ingredient to product: ", ingredientData);
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

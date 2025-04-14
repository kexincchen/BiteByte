import React, { createContext, useState, useEffect } from 'react';
import axios from 'axios';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in (e.g., from localStorage)
    const token = localStorage.getItem('token');
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      // Fetch current user info
      // In a real app, you would verify the token with the server
      const userData = JSON.parse(localStorage.getItem('user'));
      if (userData) {
        setCurrentUser(userData);
      }
    }
    setLoading(false);
  }, []);

  const login = async (email, password) => {
    try {
      // In a real app, this would be a real API call
      // For demo, simulating API response
      const response = {
        data: {
          user: {
            id: 1,
            username: 'demoUser',
            email: email,
            role: 'customer'
          },
          token: 'sample-jwt-token'
        }
      };
      
      // Would normally be: const response = await axios.post('/api/auth/login', { email, password });
      
      const { user, token } = response.data;
      
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      setCurrentUser(user);
      return user;
    } catch (error) {
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  };

  const register = async (userData) => {
    try {
      // In a real app, this would be a real API call
      // const response = await axios.post('/api/auth/register', userData);
      
      // For demo, simulating API response
      const response = {
        data: {
          id: 1,
          username: userData.username,
          email: userData.email,
          role: userData.role
        }
      };
      
      return response.data;
    } catch (error) {
      throw new Error(error.response?.data?.error || 'Registration failed');
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    delete axios.defaults.headers.common['Authorization'];
    setCurrentUser(null);
  };

  const value = {
    currentUser,
    login,
    register,
    logout,
    loading
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}; 
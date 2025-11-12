// API client for backend communication
// This file handles all HTTP requests to the web-server

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

// Helper function to get auth token from localStorage
const getToken = () => localStorage.getItem('authToken');

// Helper function to save auth token to localStorage
const saveToken = (token) => localStorage.setItem('authToken', token);

// Helper function to remove auth token from localStorage
export const clearToken = () => localStorage.removeItem('authToken');

// Helper function to check if user is authenticated
export const isAuthenticated = () => !!getToken();

// Register a new user
export const register = async (email, password) => {
  const response = await fetch(`${API_BASE_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Registration failed');
  }

  const data = await response.json();
  saveToken(data.token);
  return data;
};

// Login an existing user
export const login = async (email, password) => {
  const response = await fetch(`${API_BASE_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  const data = await response.json();
  saveToken(data.token);
  return data;
};

// Solve a math problem
export const solve = async (problem) => {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE_URL}/solve`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ problem }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to solve problem');
  }

  return response.json();
};

// Get user's problem history
export const getHistory = async () => {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE_URL}/history`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to fetch history');
  }

  return response.json();
};

// Utility functions for validation and helpers
// Pure functions with no side effects

// Validate email format
export const validateEmail = (email) => {
  if (!email) {
    return 'Email is required';
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return 'Invalid email format';
  }
  return null;
};

// Validate password
export const validatePassword = (password) => {
  if (!password) {
    return 'Password is required';
  }
  if (password.length < 6) {
    return 'Password must be at least 6 characters';
  }
  return null;
};

// Validate math problem
export const validateProblem = (problem) => {
  if (!problem || problem.trim() === '') {
    return 'Please enter a math problem';
  }
  const validChars = /^[0-9+\-*/().\s]+$/;
  if (!validChars.test(problem)) {
    return 'Problem contains invalid characters';
  }
  return null;
};

// Format date for display
export const formatDate = (dateString) => {
  const date = new Date(dateString);
  return date.toLocaleString();
};

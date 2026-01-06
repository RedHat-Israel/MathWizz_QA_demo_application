// Unit tests for utility functions
// Tests pure, non-UI helper functions in isolation

import {
  validateEmail,
  validatePassword,
  validateProblem,
  formatDate,
} from '../utils';

describe('Utils - validateEmail', () => {
  test('returns null for valid email', () => {
    expect(validateEmail('test@example.com')).toBeNull();
    expect(validateEmail('user.name@domain.co.uk')).toBeNull();
  });

  test('returns error for invalid email', () => {
    expect(validateEmail('invalid')).toBe('Invalid email format');
    expect(validateEmail('test@')).toBe('Invalid email format');
    expect(validateEmail('@example.com')).toBe('Invalid email format');
  });

  test('returns error for empty email', () => {
    expect(validateEmail('')).toBe('Email is required');
    expect(validateEmail(null)).toBe('Email is required');
    expect(validateEmail(undefined)).toBe('Email is required');
  });
});

describe('Utils - validatePassword', () => {
  test('returns null for valid password', () => {
    expect(validatePassword('password123')).toBeNull();
    expect(validatePassword('123456')).toBeNull();
  });

  test('returns error for short password', () => {
    expect(validatePassword('12345')).toBe('Password must be at least 6 characters');
    expect(validatePassword('abc')).toBe('Password must be at least 6 characters');
  });

  test('returns error for empty password', () => {
    expect(validatePassword('')).toBe('Password is required');
    expect(validatePassword(null)).toBe('Password is required');
    expect(validatePassword(undefined)).toBe('Password is required');
  });
});

describe('Utils - validateProblem', () => {
  test('returns null for valid math problem', () => {
    expect(validateProblem('2+2')).toBeNull();
    expect(validateProblem('5*10')).toBeNull();
    expect(validateProblem('(10+5)*2')).toBeNull();
    expect(validateProblem('100 - 50')).toBeNull();
  });

  test('returns error for invalid characters', () => {
    expect(validateProblem('abc')).toBe('Problem contains invalid characters');
    expect(validateProblem('2+2 hello')).toBe('Problem contains invalid characters');
    expect(validateProblem('x+y')).toBe('Problem contains invalid characters');
  });

  test('returns error for empty problem', () => {
    expect(validateProblem('')).toBe('Please enter a math problem');
    expect(validateProblem('   ')).toBe('Please enter a math problem');
    expect(validateProblem(null)).toBe('Please enter a math problem');
  });
});

describe('Utils - formatDate', () => {
  test('formats date string to locale string', () => {
    const dateString = '2024-01-15T10:30:00Z';
    const result = formatDate(dateString);
    expect(result).toContain('2024');
    expect(typeof result).toBe('string');
  });

  test('handles invalid date gracefully', () => {
    const result = formatDate('invalid-date');
    expect(result).toBe('Invalid Date');
  });
});

// Styled input component for login and register pages
// Simple form-based design with transparent buttons

import React from 'react';
import '../styles/global.css';

const LoginInputComponent = ({ email, password, onEmailChange, onPasswordChange, onSubmit, buttonText, error }) => {
  const containerStyle = {
    background: 'rgba(22, 33, 62, 0.8)',
    border: '4px solid #0f3460',
    padding: '30px',
    boxShadow: '8px 8px 0px #0f3460',
    maxWidth: '400px',
    margin: '0 auto',
    borderRadius: '0',
  };

  const inputStyle = {
    background: 'rgba(15, 52, 96, 0.9)',
    border: '3px solid #0f3460',
    padding: '15px',
    margin: '10px 0',
    width: '100%',
    color: '#eee',
    fontSize: '16px',
    fontFamily: 'Courier New, monospace',
  };

  const submitButtonStyle = {
    background: 'transparent',
    border: '4px solid rgb(249, 229, 82)',
    color: 'rgb(249, 229, 82)',
    padding: '12px 24px',
    width: '100%',
    marginTop: '10px',
    fontFamily: 'Courier New, monospace',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    borderRadius: '0',
  };

  return (
    <div style={containerStyle}>
      <form onSubmit={onSubmit}>
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => onEmailChange(e.target.value)}
          style={inputStyle}
          required
          data-testid="email-input"
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => onPasswordChange(e.target.value)}
          style={inputStyle}
          required
          data-testid="password-input"
        />

        {error && <div className="error-message">{error}</div>}

        <button
          type="submit"
          style={submitButtonStyle}
          data-testid="submit-button"
        >
          {buttonText}
        </button>
      </form>
    </div>
  );
};

export default LoginInputComponent;

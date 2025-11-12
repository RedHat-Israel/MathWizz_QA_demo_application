// Login page component
// Uses LoginInputComponent for consistent styling

import React, { useState } from 'react';
import { login } from '../api';
import { validateEmail } from '../utils';
import LoginInputComponent from './LoginInputComponent';

const LoginPage = ({ onLoginSuccess }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const emailError = validateEmail(email);
    if (emailError) {
      setError(emailError);
      return;
    }

    if (!password) {
      setError('Password is required');
      return;
    }

    try {
      await login(email, password);
      onLoginSuccess();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="page-container">
      <LoginInputComponent
        email={email}
        password={password}
        onEmailChange={setEmail}
        onPasswordChange={setPassword}
        onSubmit={handleSubmit}
        buttonText="LOGIN"
        error={error}
      />
    </div>
  );
};

export default LoginPage;

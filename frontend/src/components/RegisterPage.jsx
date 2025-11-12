// Registration page component
// Uses LoginInputComponent for consistent styling

import React, { useState } from 'react';
import { register } from '../api';
import { validateEmail, validatePassword } from '../utils';
import LoginInputComponent from './LoginInputComponent';

const RegisterPage = ({ onRegisterSuccess }) => {
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

    const passwordError = validatePassword(password);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    try {
      await register(email, password);
      onRegisterSuccess();
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
        buttonText="CREATE ACCOUNT"
        error={error}
      />
    </div>
  );
};

export default RegisterPage;

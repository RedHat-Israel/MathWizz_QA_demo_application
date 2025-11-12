// Main App component with simple routing
// Manages authentication state and page navigation

import React, { useState, useEffect } from 'react';
import { isAuthenticated, clearToken } from './api';
import RegisterPage from './components/RegisterPage';
import LoginPage from './components/LoginPage';
import SolverPage from './components/SolverPage';
import HistoryPage from './components/HistoryPage';
import './styles/global.css';

const App = () => {
  const [currentPage, setCurrentPage] = useState('login');
  const [authenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    if (isAuthenticated()) {
      setAuthenticated(true);
      setCurrentPage('solve');
    }
  }, []);

  const handleLoginSuccess = () => {
    setAuthenticated(true);
    setCurrentPage('solve');
  };

  const handleRegisterSuccess = () => {
    setAuthenticated(true);
    setCurrentPage('solve');
  };

  const handleLogout = () => {
    clearToken();
    setAuthenticated(false);
    setCurrentPage('login');
  };

  const renderPage = () => {
    if (!authenticated) {
      switch (currentPage) {
        case 'register':
          return <RegisterPage onRegisterSuccess={handleRegisterSuccess} />;
        case 'login':
        default:
          return <LoginPage onLoginSuccess={handleLoginSuccess} />;
      }
    }

    switch (currentPage) {
      case 'solve':
        return <SolverPage />;
      case 'history':
        return <HistoryPage />;
      default:
        return <SolverPage />;
    }
  };

  const appStyle = {
    backgroundImage: 'url(/frontend_background.jpg)',
    backgroundSize: 'contain',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
    minHeight: '100vh',
    width: '100vw',
  };

  const loginButtonStyle = {
    background: 'transparent',
    border: '4px solid #4da6ff',
    borderRadius: '0',
    color: '#4da6ff',
    padding: '12px 40px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    fontWeight: 'bold',
    cursor: 'pointer',
  };

  const registerButtonStyle = {
    background: 'transparent',
    border: '4px solid rgb(59, 218, 72)',
    borderRadius: '0',
    color: 'rgb(59, 218, 72)',
    padding: '12px 40px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    fontWeight: 'bold',
    cursor: 'pointer',
  };

  const solverButtonStyle = {
    background: 'transparent',
    border: '4px solid #4da6ff',
    borderRadius: '0',
    color: '#4da6ff',
    padding: '12px 40px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    fontWeight: 'bold',
    cursor: 'pointer',
  };

  const historyButtonStyle = {
    background: 'transparent',
    border: '4px solid rgb(59, 218, 72)',
    borderRadius: '0',
    color: 'rgb(59, 218, 72)',
    padding: '12px 40px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    fontWeight: 'bold',
    cursor: 'pointer',
  };

  const logoutButtonStyle = {
    background: 'transparent',
    border: '4px solid rgb(249, 229, 82)',
    borderRadius: '0',
    color: 'rgb(249, 229, 82)',
    padding: '12px 40px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    fontWeight: 'bold',
    cursor: 'pointer',
  };

  return (
    <div className="app-container" style={appStyle}>
      {!authenticated ? (
        <div className="nav-container">
          <button
            onClick={() => setCurrentPage('login')}
            style={loginButtonStyle}
          >
            LOGIN
          </button>
          <button
            onClick={() => setCurrentPage('register')}
            style={registerButtonStyle}
          >
            REGISTER
          </button>
        </div>
      ) : (
        <div className="nav-container">
          <button
            onClick={() => setCurrentPage('solve')}
            style={solverButtonStyle}
          >
            SOLVER
          </button>
          <button
            onClick={() => setCurrentPage('history')}
            style={historyButtonStyle}
          >
            HISTORY
          </button>
          <button
            onClick={handleLogout}
            style={logoutButtonStyle}
          >
            LOGOUT
          </button>
        </div>
      )}

      {renderPage()}
    </div>
  );
};

export default App;

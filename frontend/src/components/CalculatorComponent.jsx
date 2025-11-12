// Styled calculator component for the solver page
// Simple design with transparent buttons

import React from 'react';
import '../styles/global.css';

const CalculatorComponent = ({ problem, answer, onProblemChange, onSolve, loading, error }) => {
  const calculatorStyle = {
    background: 'rgba(22, 33, 62, 0.8)',
    border: '5px solid #0f3460',
    padding: '20px',
    boxShadow: '10px 10px 0px rgba(0, 0, 0, 0.5)',
    maxWidth: '500px',
    margin: '0 auto',
  };

  const screenStyle = {
    background: 'rgba(15, 52, 96, 0.9)',
    border: '3px solid #0f3460',
    padding: '20px',
    minHeight: '80px',
    marginBottom: '20px',
    fontFamily: 'Courier New, monospace',
    fontSize: '24px',
    color: '#4ecca3',
    textAlign: 'right',
    wordBreak: 'break-all',
  };

  const inputAreaStyle = {
    background: 'rgba(15, 52, 96, 0.9)',
    border: '3px solid #0f3460',
    padding: '15px',
    marginBottom: '15px',
  };

  const solveButtonStyle = {
    background: 'transparent',
    border: '4px solid rgb(249, 229, 82)',
    color: 'rgb(249, 229, 82)',
    padding: '12px 24px',
    width: '100%',
    fontFamily: 'Courier New, monospace',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    borderRadius: '0',
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSolve();
  };

  return (
    <div style={calculatorStyle}>
      <div style={screenStyle} data-testid="answer-display">
        {answer ? `= ${answer}` : 'Ready...'}
      </div>

      <form onSubmit={handleSubmit}>
        <div style={inputAreaStyle}>
          <input
            type="text"
            placeholder="Enter problem (e.g., 5*10)"
            value={problem}
            onChange={(e) => onProblemChange(e.target.value)}
            className="pixel-input"
            style={{ marginBottom: '0' }}
            disabled={loading}
            data-testid="problem-input"
          />
        </div>

        {error && <div className="error-message">{error}</div>}

        <button
          type="submit"
          style={solveButtonStyle}
          disabled={loading}
          data-testid="solve-button"
        >
          {loading ? 'SOLVING...' : 'SOLVE'}
        </button>
      </form>

      <div style={{ marginTop: '15px', fontSize: '12px', color: '#999', textAlign: 'center' }}>
        Supports: + - * / ( )
      </div>
    </div>
  );
};

export default CalculatorComponent;

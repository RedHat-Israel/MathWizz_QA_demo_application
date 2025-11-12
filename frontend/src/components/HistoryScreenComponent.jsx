// Styled history screen component for displaying problem history
// Simple design with transparent background

import React from 'react';
import '../styles/global.css';
import { formatDate } from '../utils';

const HistoryScreenComponent = ({ history, loading }) => {
  const screenStyle = {
    background: 'rgba(22, 33, 62, 0.8)',
    border: '6px solid #0f3460',
    padding: '30px',
    boxShadow: '0px 0px 20px rgba(233, 69, 96, 0.3)',
    maxWidth: '700px',
    margin: '0 auto',
    minHeight: '400px',
  };

  const headerStyle = {
    color: '#4ecca3',
    borderBottom: '2px solid #0f3460',
    paddingBottom: '10px',
    marginBottom: '20px',
    fontFamily: 'Courier New, monospace',
    fontSize: '18px',
  };

  const listStyle = {
    listStyle: 'none',
    padding: 0,
  };

  const itemStyle = {
    background: 'rgba(15, 52, 96, 0.7)',
    border: '2px solid #0f3460',
    padding: '15px',
    marginBottom: '10px',
    fontFamily: 'Courier New, monospace',
    color: '#eee',
  };

  const problemStyle = {
    color: '#4ecca3',
    fontSize: '18px',
    fontWeight: 'bold',
    marginBottom: '5px',
  };

  const timestampStyle = {
    color: '#999',
    fontSize: '12px',
    marginTop: '5px',
  };

  return (
    <div style={screenStyle}>
      <div style={headerStyle}>
        &gt; HISTORY LOG _ MATHWIZZ v1.0
      </div>

      {loading ? (
        <div className="loading">Loading history...</div>
      ) : history.length === 0 ? (
        <div style={{ color: '#999', textAlign: 'center', padding: '40px' }}>
          No history yet. Solve some problems to get started!
        </div>
      ) : (
        <ul style={listStyle} data-testid="history-list">
          {history.map((item, index) => (
            <li key={item.id || index} style={itemStyle} data-testid="history-item">
              <div style={problemStyle}>
                {item.problem} = {item.answer}
              </div>
              <div style={timestampStyle}>
                {formatDate(item.created_at)}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default HistoryScreenComponent;

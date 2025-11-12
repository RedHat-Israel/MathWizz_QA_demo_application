// Component tests for HistoryPage
// Tests UI rendering based on props using React Testing Library

import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import HistoryPage from '../components/HistoryPage';

describe('HistoryPage Component Tests', () => {
  test('renders a list based on mock history prop', () => {
    const mockHistory = [
      { id: 1, problem: '1+1', answer: '2', created_at: '2024-01-01T00:00:00Z' },
      { id: 2, problem: '5*10', answer: '50', created_at: '2024-01-02T00:00:00Z' },
      { id: 3, problem: '100-25', answer: '75', created_at: '2024-01-03T00:00:00Z' },
    ];

    render(<HistoryPage mockHistory={mockHistory} />);

    const historyItems = screen.getAllByTestId('history-item');
    expect(historyItems).toHaveLength(3);

    expect(screen.getByText('1+1 = 2')).toBeInTheDocument();
    expect(screen.getByText('5*10 = 50')).toBeInTheDocument();
    expect(screen.getByText('100-25 = 75')).toBeInTheDocument();
  });

  test('renders empty state when no history items', () => {
    render(<HistoryPage mockHistory={[]} />);

    expect(screen.getByText(/No history yet/i)).toBeInTheDocument();
    expect(screen.queryByTestId('history-item')).not.toBeInTheDocument();
  });

  test('renders correct number of items for different array sizes', () => {
    const oneItem = [
      { id: 1, problem: '2+2', answer: '4', created_at: '2024-01-01T00:00:00Z' },
    ];

    const { rerender } = render(<HistoryPage mockHistory={oneItem} />);
    expect(screen.getAllByTestId('history-item')).toHaveLength(1);

    const fiveItems = [
      { id: 1, problem: '1+1', answer: '2', created_at: '2024-01-01T00:00:00Z' },
      { id: 2, problem: '2+2', answer: '4', created_at: '2024-01-01T00:00:00Z' },
      { id: 3, problem: '3+3', answer: '6', created_at: '2024-01-01T00:00:00Z' },
      { id: 4, problem: '4+4', answer: '8', created_at: '2024-01-01T00:00:00Z' },
      { id: 5, problem: '5+5', answer: '10', created_at: '2024-01-01T00:00:00Z' },
    ];

    rerender(<HistoryPage mockHistory={fiveItems} />);
    expect(screen.getAllByTestId('history-item')).toHaveLength(5);
  });

  test('displays refresh button when history exists', () => {
    const mockHistory = [
      { id: 1, problem: '10+10', answer: '20', created_at: '2024-01-01T00:00:00Z' },
    ];

    render(<HistoryPage mockHistory={mockHistory} />);
    expect(screen.getByTestId('refresh-button')).toBeInTheDocument();
  });

  test('does not display refresh button when history is empty', () => {
    render(<HistoryPage mockHistory={[]} />);
    expect(screen.queryByTestId('refresh-button')).not.toBeInTheDocument();
  });
});

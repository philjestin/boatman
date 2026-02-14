import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { ErrorBoundary } from './components/ErrorBoundary';
import './index.css';

console.log('Main.tsx loading...');

const container = document.getElementById('root');
if (!container) {
  console.error('Root container not found!');
} else {
  console.log('Root container found, creating React root...');
  const root = createRoot(container);
  root.render(
    <React.StrictMode>
      <ErrorBoundary>
        <App />
      </ErrorBoundary>
    </React.StrictMode>
  );
  console.log('React render called');
}

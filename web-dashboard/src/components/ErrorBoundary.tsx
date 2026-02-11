import { Component } from 'react';
import type { ReactNode } from 'react'
import type { ErrorInfo } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

/**
 * ErrorBoundary catches JavaScript errors anywhere in its child component
 * tree, logs the error, and displays a fallback UI instead of a blank screen.
 *
 * Why this is critical:
 * Without this, a single thrown error in any React component (e.g. a null
 * reference in AnomalyAlerts) would crash the ENTIRE dashboard, showing
 * a white screen. With ErrorBoundary, only the broken section fails
 * while the rest of the dashboard stays interactive.
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ errorInfo });
    // In production, send this to Sentry/Datadog
    console.error('[ErrorBoundary] Caught error:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '2rem',
          minHeight: '200px',
          background: 'linear-gradient(135deg, #fef2f2, #fff)',
          border: '1px solid #fecaca',
          borderRadius: '16px',
          margin: '1rem',
        }}>
          <AlertTriangle style={{ width: 48, height: 48, color: '#ef4444', marginBottom: '1rem' }} />
          <h3 style={{ fontSize: '1.125rem', fontWeight: 600, color: '#991b1b', marginBottom: '0.5rem' }}>
            Bir şeyler yanlış gitti
          </h3>
          <p style={{ fontSize: '0.875rem', color: '#b91c1c', marginBottom: '1rem', textAlign: 'center', maxWidth: '400px' }}>
            Bu bileşen beklenmeyen bir hata ile karşılaştı.
            {this.state.error && (
              <span style={{ display: 'block', marginTop: '0.5rem', fontFamily: 'monospace', fontSize: '0.75rem', color: '#dc2626' }}>
                {this.state.error.message}
              </span>
            )}
          </p>
          <button
            onClick={this.handleRetry}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.5rem 1rem',
              background: '#ef4444',
              color: '#fff',
              border: 'none',
              borderRadius: '8px',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: 500,
              transition: 'background 0.2s',
            }}
            onMouseOver={(e) => (e.currentTarget.style.background = '#dc2626')}
            onMouseOut={(e) => (e.currentTarget.style.background = '#ef4444')}
          >
            <RefreshCw style={{ width: 16, height: 16 }} />
            Tekrar Dene
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

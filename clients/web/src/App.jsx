import { Component, Suspense } from "react";
import { Outlet } from "react-router";

class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return <>{this.state.error.toString()}</>;
    }
    return this.props.children;
  }
}

export default function App() {
  return (
    <ErrorBoundary>
      <Suspense fallback="Loading page...">
        <Outlet />
      </Suspense>
    </ErrorBoundary>
  );
}

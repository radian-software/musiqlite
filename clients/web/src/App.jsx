import { Component, Suspense } from "react";
import { Outlet } from "react-router";
import Player from "./Player";

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
        <div className="page">
          <Outlet />
          <div className="page-footer">
            <Player />
          </div>
        </div>
      </Suspense>
    </ErrorBoundary>
  );
}

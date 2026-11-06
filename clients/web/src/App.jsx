import { Component, Suspense } from "react";
import Player from "./Player";
import { Link, useLocation } from "react-router";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faArrowUp,
  faCircleChevronUp,
} from "@fortawesome/free-solid-svg-icons";

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

export default function App({ page }) {
  const location = useLocation();
  return (
    <div className="page">
      <div className="page-header">
        {location.pathname !== "/" && (
          <Link
            to={location.pathname.replace(/\/[^\/]+\/?$/, "")}
            className="page-header-navigate-up-button"
          >
            <FontAwesomeIcon icon={faCircleChevronUp} />
          </Link>
        )}
        MusiQLite &ndash; <page.Title />
      </div>
      <div className="page-content">
        <ErrorBoundary>
          <Suspense fallback="Loading page...">
            <page.Content />
          </Suspense>
        </ErrorBoundary>
      </div>
      <div className="page-footer">
        <Player />
      </div>
    </div>
  );
}

import { Suspense } from "react";
import { Outlet } from "react-router";

export default function App() {
  return (
    <Suspense fallback="Loading page...">
      <Outlet />
    </Suspense>
  );
}

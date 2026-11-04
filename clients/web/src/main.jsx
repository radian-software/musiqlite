import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router";

import AlbumsPage from "./pages/Albums.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/">
          <Route index element={<AlbumsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router";

import AlbumsPage from "./pages/Albums.jsx";
import AlbumPage from "./pages/Album.jsx";
import App from "./App.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />}>
          <Route index element={<AlbumsPage />} />
          <Route path="album/:albumName" element={<AlbumPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);

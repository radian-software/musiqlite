import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router";

import AlbumsPage from "./pages/Albums.jsx";
import AlbumPage from "./pages/Album.jsx";
import HomePage from "./pages/Home.jsx";
import PlaylistsPage from "./pages/Playlists.jsx";
import App from "./App.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route index element={<App page={HomePage} />} />
        <Route path="albums" element={<App page={AlbumsPage} />} />
        <Route path="album/:albumName" element={<App page={AlbumPage} />} />
        <Route path="playlists" element={<App page={PlaylistsPage} />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);

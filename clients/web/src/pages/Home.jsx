import { Link } from "react-router";

export default {
  Title: function HomePageTitle() {
    return "Home";
  },
  Content: function HomePage() {
    return (
      <>
        <p>Welcome, you may want:</p>
        <ul>
          <li>
            <Link to="/albums">Albums</Link>
          </li>
          <li>
            <Link to="/playlists">Playlists</Link>
          </li>
        </ul>
      </>
    );
  },
};

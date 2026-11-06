import { Link } from "react-router";
import { useFetchJSON } from "../util";

export default {
  parentPath: "/",
  Title: function PlaylistsPageTitle() {
    return "Playlists";
  },
  Content: function PlaylistsPage() {
    const playlists = useFetchJSON("/api/v0/playlists");
    return (
      <div className="list-rows">
        {playlists.map((playlistData) => {
          return (
            <div className="list-row" key={playlistData.name}>
              <Link to={`/playlist/${playlistData.name}`}>
                {playlistData.name}
              </Link>
            </div>
          );
        })}
      </div>
    );
  },
};

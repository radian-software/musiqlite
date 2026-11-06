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
              <div>{playlistData.name}</div>
            </div>
          );
        })}
      </div>
    );
  },
};

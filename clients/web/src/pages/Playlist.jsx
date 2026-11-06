import { useParams } from "react-router";

import {
  fetchJSON,
  formatDuration,
  parseBlob,
  useErrorHandling,
  useFetchJSON,
} from "../util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faPlay } from "@fortawesome/free-solid-svg-icons";

export default {
  parentPath: "/playlists",
  Title: function PlaylistPageTitle() {
    const { playlistName } = useParams();
    return `Playlist: ${playlistName}`;
  },
  Content: function PlaylistPage() {
    const handleErrors = useErrorHandling();
    const { playlistName } = useParams();
    const songs = useFetchJSON("/api/v0/playlist", {
      name: playlistName,
      tag: ["oid", "song", "track", "audio"],
    });
    return (
      <>
        <button
          className="page-play-button"
          onClick={handleErrors(async () => {
            await fetchJSON("PUT /api/v0/playlist", {
              name: "Up Next",
              song: songs.map((songData) => songData.oid),
            });
          })}
        >
          <FontAwesomeIcon icon={faPlay} />
        </button>
        <div className="list-rows">
          {songs.map((songData, idx) => {
            const audio = parseBlob(songData.audio);
            const duration = formatDuration(audio.props.duration);
            // This is reusing the same CSS of the album song list, but
            // should be replaced with a different layout instead. The
            // new layout should be two-line to show the album and maybe
            // artist on the second line, and should include album art
            // thumbnails.
            return (
              <div
                className="list-row album-song-list-row"
                key={songData.entry_oid}
              >
                <div className="album-song-list-row-track">{idx + 1}</div>
                <div className="album-song-list-row-song">{songData.song}</div>
                <div className="album-song-list-row-duration">{duration}</div>
              </div>
            );
          })}
        </div>
      </>
    );
  },
};

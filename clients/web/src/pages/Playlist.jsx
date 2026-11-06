import { useParams } from "react-router";

import { formatDuration, parseBlob, useFetchJSON } from "../util";

export default {
  parentPath: "/playlists",
  Title: function PlaylistPageTitle() {
    const { playlistName } = useParams();
    return `Playlist: ${playlistName}`;
  },
  Content: function PlaylistPage() {
    const { playlistName } = useParams();
    const songs = useFetchJSON("/api/v0/playlist", {
      name: playlistName,
      tag: ["oid", "song", "track", "audio"],
    });
    return (
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
            <div className="list-row album-song-list-row" key={songData.oid}>
              <div className="album-song-list-row-track">{idx + 1}</div>
              <div className="album-song-list-row-song">{songData.song}</div>
              <div className="album-song-list-row-duration">{duration}</div>
            </div>
          );
        })}
      </div>
    );
  },
};

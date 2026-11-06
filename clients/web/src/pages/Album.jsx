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
  parentPath: "/albums",
  Title: function AlbumPageTitle() {
    const { albumName } = useParams();
    return `Album: ${albumName}`;
  },
  Content: function AlbumPage() {
    const handleErrors = useErrorHandling();
    const { albumName } = useParams();
    const songs = useFetchJSON("/api/v0/songs", {
      "filter-album": albumName,
      tag: ["oid", "song", "track", "audio"],
    });
    return (
      <>
        <button
          className="page-play-button"
          onClick={handleErrors(async () => {
            await fetchJSON("PUT /api/v0/playlist", {
              name: "Up Next",
              is_upnext: true,
              song: songs.map((songData) => songData.oid),
            });
          })}
        >
          <FontAwesomeIcon icon={faPlay} />
        </button>
        <div className="list-rows">
          {songs.map((songData) => {
            const audio = parseBlob(songData.audio);
            const duration = formatDuration(audio.props.duration);
            return (
              <div className="list-row album-song-list-row" key={songData.oid}>
                <div className="album-song-list-row-track">
                  {songData.track}
                </div>
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

import { useParams } from "react-router";

import { formatDuration, parseBlob, useFetchJSON } from "../util";

export default function AlbumPage() {
  const { albumName } = useParams();
  const songs = useFetchJSON("/api/v0/songs", {
    "filter-album": albumName,
    tag: ["oid", "song", "track", "audio"],
  });
  return (
    <>
      <div className="page-header">MusiQLite - Album: {albumName}</div>
      <div className="page-content">
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
      </div>
    </>
  );
}

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
      <div className="list-rows">
        {songs.map((songData) => {
          const audio = parseBlob(songData.audio);
          const duration = formatDuration(audio.props.duration);
          return (
            <div className="list-row" key={songData.oid}>
              {songData.track} {songData.song} {duration}
            </div>
          );
        })}
      </div>
    </>
  );
}

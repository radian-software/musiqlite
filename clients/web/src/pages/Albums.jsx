import { Link } from "react-router";

import { parseBlob, useFetchJSON } from "../util";

export default {
  Title: function AlbumsPageTitle() {
    return "Albums";
  },
  Content: function AlbumsPage() {
    const albums = useFetchJSON("/api/v0/groups", {
      partition: "album",
      sort: "album",
      tag: ["album_artist", "artwork"],
    });
    return (
      <div className="cards">
        {albums.map((albumData) => {
          const albumName = albumData.partition_tag_value;
          const artwork = parseBlob(
            albumData.tag_values.artwork.most_common_value,
          );
          // Thumbnails are always jpeg
          const artworkFilename = `${artwork.digest}.jpg?w=300`;
          return (
            <Link to={`/album/${albumName}`} key={albumName}>
              <div className="card">
                <img
                  className="card-picture"
                  src={`http://localhost:8080/api/v0/media/${artworkFilename}`}
                />
                <div className="card-caption">{albumName}</div>
              </div>
            </Link>
          );
        })}
      </div>
    );
  },
};

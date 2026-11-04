import { useEffect, useState } from "react";
import { Link } from "react-router";

const parseBlob = (blob) => {
  const [_, size, digest, propsList] = blob.split(":");
  const props = {};
  for (const propEntry of propsList.split(",")) {
    const [key, val] = propEntry.split("=");
    props[key] = val;
  }
  return { size, digest, props };
};

export default function AlbumsPage() {
  const [albumsData, setAlbumsData] = useState(null);
  useEffect(() => {
    (async () => {
      try {
        const resp = await fetch(
          `http://localhost:8080/api/v0/groups?partition=album&sort=album&tag=album_artist&tag=artwork`,
        );
        if (!resp.ok) {
          throw new Error(`HTTP ${resp.status} ${resp.statusText}`);
        }
        setAlbumsData({ value: await resp.json() });
      } catch (error) {
        setAlbumsData({ error });
      }
    })();
  }, []);
  if (albumsData === null) return <>Loading albums...</>;
  if (albumsData.error)
    return <>Failed to load albums: {albumsData.error.toString()}</>;
  return (
    <>
      <div className="page-header">MusiQLite - Albums</div>
      <div className="cards">
        {albumsData.value.map((albumData) => {
          const albumName = albumData.partition_tag_value;
          const artwork = parseBlob(
            albumData.tag_values.artwork.most_common_value,
          );
          // Thumbnails are always jpeg
          const artworkFilename = `${artwork.digest}.jpg?w=300`;
          return (
            <Link to={`/album/${albumName}`}>
              <div className="card" key={albumName}>
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
    </>
  );
}

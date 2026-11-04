import { useEffect, useState } from "react";

const parseBlob = (blob) => {
  const [_, size, digest, propsList] = blob.split(":");
  const props = {};
  for (const propEntry of propsList.split(",")) {
    const [key, val] = propEntry.split("=");
    props[key] = val;
  }
  return { size, digest, props };
};

export default function App() {
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
          const artwork = parseBlob(
            albumData.tag_values.artwork.most_common_value,
          );
          const artworkFilename = `${artwork.digest}.${artwork.props.type}`;
          return (
            <div className="card" key={albumData.partition_tag_value}>
              <img
                className="card-picture"
                src={`http://localhost:8080/api/v0/media/${artworkFilename}`}
              />
              <div className="card-caption">
                {albumData.partition_tag_value}
              </div>
            </div>
          );
        })}
      </div>
    </>
  );
}

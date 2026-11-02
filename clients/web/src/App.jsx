import { useEffect, useState } from "react";

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
    <ul>
      {albumsData.value.map((albumData) => (
        <li key={albumData.partition_tag_value}>
          {albumData.partition_tag_value}
        </li>
      ))}
    </ul>
  );
}

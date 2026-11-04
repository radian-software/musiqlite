import { useParams } from "react-router";

export default function AlbumPage() {
  const { albumName } = useParams();
  return <>{albumName}</>;
}

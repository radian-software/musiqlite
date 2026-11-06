import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faPlay,
  faBackwardStep,
  faForwardStep,
} from "@fortawesome/free-solid-svg-icons";
import { useFetchJSON } from "./util";

export default function Player() {
  const playlist = useFetchJSON(
    "/api/v0/playlist",
    {
      name: "Up Next",
      tag: ["song", "audio"],
    },
    { knownErrorCodes: ["no_such_playlist"] },
  );
  return (
    <>
      <button className="player-button player-button-skip-to-previous">
        <FontAwesomeIcon icon={faBackwardStep} />
      </button>
      <button className="player-button player-button-play-pause">
        <FontAwesomeIcon icon={faPlay} />
      </button>
      <button className="player-button player-button-skip-to-next">
        <FontAwesomeIcon icon={faForwardStep} />
      </button>
      {playlist.error_code !== "no_such_playlist" && playlist.is_upnext && (
        <div>Playback position: {playlist.playback_position}</div>
      )}
    </>
  );
}

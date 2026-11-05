import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faPlay,
  faBackwardStep,
  faForwardStep,
} from "@fortawesome/free-solid-svg-icons";

export default function Player() {
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
    </>
  );
}

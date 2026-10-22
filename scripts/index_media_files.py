#!/usr/bin/env python3

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess


def walk_tree(path: Path):
    yield path
    if path.is_dir():
        for child in sorted(path.iterdir()):
            yield from walk_tree(child)


def get_image_geometry(image_path: Path):
    output = subprocess.run(
        ["magick", "identify", str(image_path)],
        stdout=subprocess.PIPE,
        encoding="utf-8",
        check=True,
    ).stdout
    _, geometry, *_ = output.removeprefix(str(image_path)).split()
    width, height = geometry.split("x")
    return int(width), int(height)


def get_audio_duration(audio_path: Path):
    output = subprocess.run(
        [
            "ffprobe",
            "-i",
            str(audio_path),
            "-show_entries",
            "format=duration",
            "-v",
            "quiet",
            "-output_format",
            "json",
        ],
        stdout=subprocess.PIPE,
        encoding="utf-8",
        check=True,
    ).stdout
    data = json.loads(output)
    return float(data["format"]["duration"])


parser = argparse.ArgumentParser()
parser.add_argument("output_file")
parser.add_argument("input_dirs", nargs="+")
parser.add_argument("--start-from", default="")
args = parser.parse_args()

output_file = Path(args.output_file).resolve()
input_dirs = [Path(p).resolve() for p in args.input_dirs]
start_from_path = Path(args.start_from).resolve() if args.start_from else None

if not output_file.parent.is_dir():
    raise FileNotFoundError(output_file.parent)

try:
    with open(output_file) as f:
        file_index = json.load(f)
except FileNotFoundError:
    file_index = {}

try:
    has_started = start_from_path is None
    for input_dir in input_dirs:
        for path in walk_tree(input_dir):
            if not has_started:
                if path != start_from_path:
                    continue
                has_started = True
            if not path.is_file():
                if path.is_dir():
                    print(f"indexing: {path} ...")
                continue
            with open(path, "rb") as f:
                hasher = hashlib.sha256(usedforsecurity=False)
                while chunk := f.read(1024 * 1024):
                    hasher.update(chunk)
            digest = hasher.hexdigest()
            info = {
                "size": path.stat().st_size,
                "digest": digest,
            }
            match "".join(path.suffixes):
                case ".png" | ".jpg" | ".jpeg":
                    width, height = get_image_geometry(path)
                    info |= {
                        "width": width,
                        "height": height,
                    }
                case ".mp3":
                    info |= {"duration": get_audio_duration(path)}
                case suffix:
                    raise RuntimeError(f"Unsupported file format {suffix} in {path}")
            file_index[str(path.relative_to(input_dir))] = info
finally:
    with open(f"{output_file}.tmp", "w") as f:
        json.dump(file_index, f, indent=2)
        f.write("\n")
    Path(f"{output_file}.tmp").rename(output_file)

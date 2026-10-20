#!/usr/bin/env python3

import argparse
import hashlib
from pathlib import Path


def walk_tree(path: Path):
    yield path
    if path.is_dir():
        for child in sorted(path.iterdir()):
            yield from walk_tree(child)


parser = argparse.ArgumentParser("import_to_object_store.py")
parser.add_argument("input_dir")
parser.add_argument("object_store_dir")
# If you specify this argument then it still scans the input dir, but
# it assumes there's an exact copy of the input dir in the reference
# dir, and it uses the reference dir to read file contents, but it
# still makes hardlinks to the input dir. This helps me maintain a
# library on a USB drive where I/O is much slower, when I also have a
# copy of the library on the main SSD. Hardlinks of course cannot be
# cross-device, though, hence the need for two dirs.
parser.add_argument("--reference-input-dir", default="")
args = parser.parse_args()

input_dir = Path(args.input_dir).resolve()
reference_input_dir = (
    Path(args.reference_input_dir) if args.reference_input_dir else None
)
object_store_dir = Path(args.object_store_dir).resolve()
objects_dir = object_store_dir / "objects"

for path in walk_tree(input_dir):
    if not path.is_file():
        if path.is_dir():
            print(f"importing: {path} ...")
        continue
    read_path = (
        reference_input_dir / path.relative_to(input_dir)
        if reference_input_dir
        else path
    )
    with open(read_path, "rb") as f:
        hasher = hashlib.sha256(usedforsecurity=False)
        while chunk := f.read(1024 * 1024):
            hasher.update(chunk)
    digest = hasher.hexdigest()
    prefix, suffix = digest[:2], digest[2:]
    object_path = objects_dir / prefix / suffix
    object_path.parent.mkdir(parents=True, exist_ok=True)
    object_path.hardlink_to(path)

print("finished.")

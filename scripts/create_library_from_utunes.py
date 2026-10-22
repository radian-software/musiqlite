#!/usr/bin/env python3

import argparse
import json
from pathlib import Path


def quote(text: str) -> str:
    return "'" + text.replace("'", "''") + "'"


def normalize_utunes_boolean(boolean: str) -> bool:
    match boolean:
        case "yes":
            return True
        case "no":
            return False
    raise RuntimeError(f"Unsupported boolean value: {boolean}")


parser = argparse.ArgumentParser("create_library_from_utunes")
parser.add_argument("utunes_library_file")
parser.add_argument("--utunes-index-file", required=True)
args = parser.parse_args()

utunes_library_file = Path(args.utunes_library_file)
with open(utunes_library_file) as f:
    utunes_library = json.load(f)

    utunes_index_file = Path(args.utunes_index_file)
with open(utunes_index_file) as f:
    utunes_content_index = json.load(f)

print("""
CREATE TABLE song (
  -- used for display
  album,
  album_artist,
  artist,
  composer,
  disc,
  track,
  song,
  year,

  -- sorting related
  album_sort,
  album_artist_sort,
  artist_sort,
  composer_sort,
  song_sort,

  -- content
  audio,
  artwork,
  duration,

  -- informational
  acquired_illegally,
  acquired_legally,
  as_bundle,
  as_gift,
  min_price,
  paid,
  date,
  source,
  refined_source,
  tracklist,
  group_id,
  utunes_id
);

CREATE TABLE playlist (
  name,
  is_upnext_playlist
);

CREATE TABLE playlist_song (
  playlist_oid,
  song_oid,
  ordering
);

CREATE TABLE listen (
  song_oid,
  playlist_oid,
  timestamp,
  device_string_oid
);

CREATE TABLE string (
  value
);
""".strip())

values = []
for song in utunes_library["songs"].values():
    get_required_string = lambda key: quote(song[key])
    get_optional_string = lambda key: (
        "NULL" if key not in song else get_required_string(key)
    )
    get_required_number = lambda key: song[key]
    get_optional_number = lambda key: (
        "NULL" if key not in song else get_required_number(key)
    )
    get_sort_string = lambda key: (
        "NULL"
        if (song.get(f"{key}_sort") in (None, song.get(key)))
        else quote(song.get(f"{key}_sort"))
    )
    get_object_string = lambda key: quote(
        f"blob:{utunes_content_index[song[key]]['size']}:{utunes_content_index[song[key]]['digest']}"
    )
    get_index_content = lambda key, prop: utunes_content_index[song[key]][prop]
    get_required_boolean = lambda key: (
        "TRUE" if normalize_utunes_boolean(song[key]) else "FALSE"
    )
    values.append(f"""
  (
    {get_required_string('album')}, -- album
    {get_required_string('album_artist')}, -- album_artist
    {get_required_string('artist')}, -- artist
    {get_optional_string('composer')}, -- composer
    {get_required_number('disc')}, -- disc
    {get_optional_number('track') or 'NULL'}, -- track
    {get_required_string('song')}, -- song
    {get_required_number('year')}, -- year

    {get_sort_string('album')}, -- album_sort
    {get_sort_string('album_artist')}, -- album_artist_sort
    {get_sort_string('artist')}, -- artist_sort
    {get_sort_string('composer')}, -- composer_sort
    {get_sort_string('song')}, -- song_sort

    {get_object_string('filename')}, -- audio
    {get_object_string('artwork')}, -- artwork
    {get_index_content('filename', 'duration')}, -- duration

    {get_required_boolean('acquired_illegally')}, -- acquired_illegally
    {get_required_boolean('acquired_legally')}, -- acquired_legally
    {get_required_boolean('as_bundle')}, -- as_bundle
    {get_required_boolean('as_gift')}, -- as_gift
    {get_optional_number('min_price')}, -- min_price
    {get_optional_number('paid')}, -- paid
    {get_required_string('date')}, -- date
    {get_required_string('source')}, -- source
    {get_optional_string('refined_source')}, -- refined_source
    {get_optional_string('tracklist')}, -- tracklist
    {get_optional_string('group_id')}, -- group_id
    {get_required_string('id')} -- utunes_id
  )
    """.strip())
print(f"""
INSERT INTO song VALUES
  {',\n  '.join(values)};
""".strip())

values = []
for playlist_name in utunes_library["playlists"]:
    values.append(f"({quote(playlist_name)}, FALSE)")
print(f"""
INSERT INTO playlist VALUES
  {',\n  '.join(values)};
""".strip())

values = []
for playlist_name, song_ids in utunes_library["playlists"].items():
    for ordering, song_id in enumerate(song_ids, start=1):
        values.append(f"({quote(playlist_name)}, {quote(song_id)}, {ordering})")
print(f"""
WITH input(playlist_name, song_id, ordering) AS (
  VALUES
  {',\n  '.join(values)}
)
INSERT INTO playlist_song
SELECT song.oid, playlist.oid, input.ordering FROM input
LEFT JOIN playlist ON playlist.name = input.playlist_name
LEFT JOIN song ON song.utunes_id = input.song_id;
""".strip())

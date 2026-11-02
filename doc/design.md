This document explains the work-in-progress design of MusiQLite.

## Client-server model

MusiQLite is first and foremost a server that has an attached file
storage for music data, with a SQLite database to keep track of
metadata, that provides access to music metadata and audio streams
through an HTTP interface.

In order to make use of MusiQLite, you use a client that connects to
the server. The server serves a default client that uses HTML, CSS,
and JavaScript. However, you can use a different client if you prefer,
for example one that translates a command-line interface into the
necessary HTTP requests.

If you want to access your media from only one device, you can run the
server and the client both on that device. If you want to access it on
multiple devices, you run the server on an internet-accessible host
that can be queried by clients running on each device where you want
to access media. However, this mode of operation requires that you be
connected to the internet whenever you want to access your media.

To support offline usage, your device can run both a client and its
own copy of the server, with the server configured as a "downstream"
server pointing at the "upstream" server running on the internet. In
this mode, by default the downstream server proxies all requests to
the upstream server. But, you can enable offline mode, which causes
the downstream server to maintain its own copy of the upstream
server's SQLite database, as well as its own local file store
(optionally containing some subset of the media stored by the upstream
server's file store). Then requests are answered using the local copy
of the database, and any media that is available in the local file
store can be played offline. The downstream server maintains some
SQLite data that is not synchronized with the upstream server, such as
the state of its own local file store, but is returned to the client
so that it can display information about downloads to the user.

The need for the server to be able to run on a variety of devices
leads to the decision for it to be implemented in Go.

## Database synchronization

When enabling offline mode on a downstream server, the downstream
server downloads a complete copy of the upstream server's SQLite
database. All devices with database copies must keep them in sync with
each other to the extent possible. This is impossible in general, due
to the CAP theorem, so compromises are made to provide availability
and partition tolerance at the expense of consistency (if two devices
make incompatible edits to the same data while both are offline, then
one of them will have their edit reverted when coming back online).

The server translates client API calls into SQLite database queries.
Whenever any query is run that mutates the database, it is recorded in
a write-ahead log, alongside a corresponding query that would undo the
change that was made, a sequence number, and additional metadata that
allows for identifying which queries may conflict with one another if
made concurrently.

The strategy for resolving conflicts is this: When a device is online,
it continuously receives queries that were executed on the upstream
database, and replays them locally. When it is offline, and a query is
made locally, it is recorded in the write-ahead log. When it comes
back online, the server is asked for queries that had been made in the
meantime. If there are any, then all queries in the local write-ahead
log are undone in reverse order, the queries from the server are
played back, and then the write-ahead log is replayed again, with
those queries being sent to the server for relaying to other devices.
If new queries had been applied on the server during this process,
then it is repeated until the race condition is resolved.

Note that the above strategy guarantees a single canonical state of
the database, which is maintained by the upstream server. All local
edits are treated as provisional until they can be synchronized
online. However, providing that conflicting edits are not made to the
same data concurrently on different devices, the above strategy also
guarantees that all edits made offline will succeed, and the
undo-replay process will be invisible to the user.

In the case that there truly are conflicts, these will emerge during
the last stage of the synchronization process, when the local
write-ahead log is replayed after the most recent edits from the
server have been applied. The local edits can still be applied, but
may change in meaning due to the impact of prior edits from the
server. To facilitate this, queries are to be written in such a way as
to maximize the chance of their still retaining the same semantic
meaning to the user even when their relative ordering changes.
However, it is accepted that this may fail in some circumstances. In
the worst case, an edit may be dropped entirely, for example if it
refers to data that no longer exists, which the user will be informed
of.

Note that SQLite provides suggestions which can aid in implementing
query undo capability: <https://www.sqlite.org/undoredo.html>

## Client API

All operations are exposed via HTTP in two forms: REST and Websocket.
Individual operations can be executed conveniently via REST, but
clients performing many sequential operations can subscribe to a
Websocket connection in order to make the same requests with lower
latency, and to be informed of changes from other clients in real
time.

Some API routes that allow for complex queries have two formats: a
full format that takes a JSON object in the query string or request
body, and a simplified format that uses individual query parameters,
and generates a corresponding JSON object according to simple rules.
This is intended to allow for simple requests to be made simply,
without the need to write code. However, the simplified format does
not necessarily support every possible query pattern.

## Database schema

The MusiQLite library is primarily a collection of songs. Each song is
a key-value mapping of metadata ("tags"), low-cardinality key strings
to varying-cardinality value strings. In addition, there are some
"virtual" tags that may be referred to by queries, but are populated
by system logic rather than the user. These are prefixed by an
underscore, which is a disallowed prefix for user-defined tags. One
such virtual tag is `_id`, which is a UUID identifying the song as a
unique object; others are discussed later.

The user interface can treat certain tags specially, based on your
configuration, for example to group songs into albums, show their
artwork, and display them in order by disc and track.

The library database keeps some special song data separate from the
main collection of key-value tag objects. The file storage is one
example of this. Other examples are:

* Playlists. These have a similar key-value tag object as songs,
  although they need not share the same schema, as well as an ordered
  list of song references (UUIDs). They are stored in a separate
  table. Playlists can be either user-defined, or defined as queries
  that return a list based on the current library contents, like a
  materialized view on a database.
* Listening history. This is a special ordered log of entries. Each
  entry has a song reference (UUID), as well as a timestamp and
  possibly an arbitrary key-value tag object similar to songs,
  although not necessarily with the same schema.
* Up next. This is a playlist that is created with a special tag to
  indicate that it is to be treated specially. The frontend interacts
  with it when queueing songs for playback. You have the option, but
  not the requirement, to copy the contents from one such playlist to
  another such playlist (perhaps on a different device), to achieve
  the effect of playback handoff between devices.
* Playback position within up next playlist and within a song.

These additional constructs may generate "virtual" tags on each song,
which are computed on demand if they are referred to from a query. For
example, `?sort=_last_listen` to sort by most recently listened. You
can also use the full query interface to pass additional parameters to
virtual tags, which may affect their computation. For example, to sort
by most recently listened to on a specific device:

```
{
  "sort": [
    {
      "conditional_tag": {
        "tag": "_last_listen",
        "conditions": [
          {
            "tag": "device",
            "value": "jupiter"
          }
        ]
      }
    }
  ]
}
```

Or for playlists, you can filter on membership, but only using unique
identifier of the playlist: `?filter-_playlist_member_<id>=yes`.

## File storage

Tag values, in any field, can also be strings of the form
`blob:<length>:<sha256sum>:<detail>` to reference a blob stored in the
file storage. So, the album artwork and audio data of a song are both
represented as tags, and the usage of a shared file storage means that
many songs with the same album artwork need not consume any additional
space to store copies of that artwork. (As an advanced use case,
multiple copies of the same song could appear under different names
and metadata, without consuming extra storage space.)

The `<detail>` field is a key-value mapping that specifies additional
information about the file content. The required key is `type`, and
other keys may be required based on the type. For example:

* `type=png,width=1400,height=1400`
* `type=mp3,duration=208.397`

## API route examples

Let's say that you open your MusiQLite homepage which is configured to
show tiles of all the albums in your library. Then, the frontend would
query:

```
GET /api/v0/groups?partition=album&sort=album_sort
```

This groups all the songs in the library by the `album` key, and for
each group, returns an object that summarizes the other values that
each tag takes within the group (with a limit of how many distinct
values are reported for a single tag).

The returned groups are sorted by their common values (one is selected
if it varies over the course of the group) of the `album_sort` tag.

You can also filter the search results, however this is not used in
the example query given, because it is showing all albums in the
entire library. However, for example if you had a page which showed
albums by a particular artist, you could add the query parameter
`filter-album_artist=C418` for example, and result groups would be
computed based on only songs that match that filter (and any others
supplied).

An object in the returned array from this API call might be:

```
{
  "partition": "Snailiad",
  "common_tags": {
    "album": "Snailiad",
    "album_sort": "Snailiad",
    "album_artist": "Auriplane",
    "album_artist_sort": "Auriplane",
    "artist": "Auriplane",
    "artist_sort": "Auriplane",
    "artwork": "blob:114992:f42009d8cf4c5b5dc5e6c00ae56f8898969ffcef2b5a348dfc14ca3c1f31d3cd",
    "composer": "Auriplane",
    "composer_sort": "Auriplane",
    "disc": "1",
    "year": "2011"
  },
  "varying_tags": {
    "_id": [
      {
        "count": 1,
        "value": "669da727-215e-4085-bd84-866e4023abf8"
      },
      {
        "count": 1,
        "value": "2c138ca7-ccf2-4270-abf8-1e3b1b84d04b"
      },
      {
        "count": 1,
        "value": "3846268f-fa9e-4130-86f8-37e2648a85d4"
      },
      {
        "count": 7
      }
    ],
    "song": [
      {
        "count": 1,
        "value": "blob:1484873:ffc266e26e09b3d1e8ec76f80d712043edba678ae78952e68e4848faa0ef5463"
      },
      {
        "count": 1,
        "value": "blob:1188915:618aab3b7dcc46628e788a0a827356c18212635174d209908bd5a0cd34b55100"
      },
      {
        "count": 1,
        "value": "blob:1246346:d6eeb42aade1139be05522bbe9d7816dc6a6604efee59c0091d0f3abfa0c178f"
      },
      {
        "count": 7
      }
    ],
    "track": [
      {
        "count": 1,
        "value": "1"
      },
      {
        "count": 1,
        "value": "2"
      },
      {
        "count": 1,
        "value": "3"
      },
      {
        "count": 7
      }
    ]
  }
}
```

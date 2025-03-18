# thrasher-music-catalog

The music catalog is a component of the Thrasher suite which provides
faceted classification (i.e. "tagging") for a music collection.

It operates over a directory tree of MP3 files, building and managing
a SQLite database which contains the data extracted from ID3 tags and
the filesystem, as well as user-supplied facets/tags for each track.

It is implemented in two parts:

- `reader`, a module which provides query operations regarding the
  current state of the database
- `writer`, a module which handles all update operations

A CLI utility is also provided, functioning as both a ready-made
management interface, and a demo of the libraries.


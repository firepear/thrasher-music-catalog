# thrasher-music-catalog

The music catalog is a component of the Thrasher music library suite
which provides faceted classification (i.e. "tagging") and metadata
management for a music collection.

It operates over a directory tree of MP3 files, building and managing
a SQLite database which contains the data extracted from ID3 tags and
the filesystem, as well as user-supplied facets/tags for each track.

It is implemented in two parts:

- `tmc`, a package which provides read-only operations to an in-memory
  copy of the catalog database
- `tmcu`, a package which handles all update operations to the on-disk
  catalog database

This split makes data safety a no-brainer, as applications which only
need to read or query the catalogue simply do not import the updater
module.

## tmctool

A CLI utility, `tmctool`, is also provided. It provides basic catalog
maintenance functions:

- Database creation
- Music collection scanning (catalog import + update)
- Applying and removing facets to catalogued tracks
- ID3 tag editing of files (because your music collection is the
  source of truth for metadata other than non-genre facets)


# thrasher-music-catalog

The catalog is the component of the Thrasher music library suite which
provides faceted classification (i.e. "tagging") and metadata
management for a music collection.

It operates over a directory tree of MP3 files, building and managing
a SQLite database (AKA the catalog) which contains the data extracted
from ID3 tags and the filesystem, as well as user-supplied facets/tags
for each track.

It is implemented in two parts:

- `tmc`, a package which provides read-only operations to an in-memory
  copy of the catalog database
  - Any instances of `tmc.Catalog` which use the same DB name will
    share memory cache, decreasing startup time and resource usage
- `tmcu`, a package which handles all update operations to the on-disk
  catalog database

This split makes data integrity easy, as applications which do not
need write access to the catalogue simply should not import the
updater package.

## tmctool

A CLI utility, `tmctool`, is also provided. It provides basic catalog
maintenance functions:

- Database creation
- Music collection scanning (catalog import + update)
- Applying and removing facets to catalogued tracks
- ID3 tag editing of files (because your music collection is the
  source of truth for metadata other than non-genre facets)


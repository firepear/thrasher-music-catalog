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

## Filter format

The catalog is queried by setting a _filter_ and then fetching
tracks. The filter is set by calling `Calendar.ParseFormat` with a
_format string_ argument. An example:

`f:funk&&((y:197%//>=1995))||a:snarky puppy\\confunktion`

That looks horrible, but the first thing to note is that whitespace is
only significant within attribute values (which we'll come to in a
moment). The second thing to note is that attributes themselves have
expanded forms. So that can be rewritten as:

`facet: funk  &&  ((year:197%  //  >=1995))  ||  artist: snarky puppy  \\  confunktion`

This looks a lot more sensible, and in fact it looks a lot like the
`WHERE` clause of a SQL query. That's exactly what it becomes. If we
examine `c.Filter`, in the middle of it is:

`WHERE facets LIKE ? AND ( year LIKE ? OR year >= ? ) OR artist LIKE ? AND artist LIKE ?`

So `&&` and `||` are the logical operators they look like, and map to
`AND` and `OR`. Doubled parens (`((` and `))`) are escapes for a
single paren in the generated SQL, and are acting as grouping for
order of operations.

Most of what's left is `attribute: value` pairs, which work exactly
the way you expect them to, except that no quoting is needed. The
supported attributes are:

- `artist` (short: `a`)
- `album` (short: `b`)
- `title` (short: `t`)
- `year` (short: `y`)
- `facets` (alt: `facet`, `f`)

You've probably noticed that `//` and `\\` _also_ map to `OR` and
`AND`. You may have noticed that they occur within attribute
values. They're syntactic sugar to compactly specify multiple values
for a single attribute.

`a:x//y` is equivalent to `a:x || a:y`

## tmctool

A CLI utility, `tmctool`, is also provided. It provides basic catalog
maintenance functions:

- Database creation
- Music collection scanning (catalog import + update)
- Applying and removing facets to catalogued tracks
- ID3 tag editing of files (because your music collection is the
  source of truth for metadata other than non-genre facets)


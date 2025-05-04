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

## Instantiate a catalog instance

`Catalog.New` takes two arguments: the path to the on-disk SQLite DB,
and the name to be used for the in-memory copy (which is the working
datastore for `Catalog`).

```
import (
    tmc "github.com/firepear/thrasher-music-catalog"
)

func main() {
    c, err := tmc.New("/path/to/database.db", "memDbName")
    if err != nil {
        // as appropriate...
    }
    // c is ready to use
)
```

## Filtering

The catalog is accessed by first setting a _filter_ and then fetching
tracks. The filter is set by calling `c.ParseFormat` with a _format
string_ argument. An example:

`c.ParseFilter("f:funk&&((y:197%//>=1995))||a:snarky puppy\\confunktion")`

That looks horrible, but the first thing to note is that whitespace is
only significant within attribute values (which we'll come to in a
moment). The second thing to note is that attributes themselves have
expanded forms. The format string can be rewritten as follows:

`facet: funk  &&  ((year:197%  //  >=1995))  ||  artist: snarky puppy  \\  confunktion`

This looks a lot more sensible, and in fact it resembles the `WHERE`
clause of a SQL query. That's exactly what it becomes. If we examine
`c.Filter`, in the middle of it is:

`WHERE facets LIKE ? AND ( year LIKE ? OR year >= ? ) OR artist LIKE ? AND artist LIKE ?`

(No, it's not the most sensical query; it's an example.)

So `&&` and `||` are the logical operators they look like, and map to
`AND` and `OR`. Doubled parens (`((` and `))`) are escapes for a
single paren in the generated SQL, and are grouping for order of
operations, as expected.

Most of what's left is `attribute: value` pairs, which work exactly
the way you expect them to, except that no quoting is needed. The
supported attributes are:

- `artist` (short: `a`)
- `album` (short: `b`)
- `title` (short: `t`)
- `year` (short: `y`)
  - Only `year` supports prefacing a value with `<=`, `>=`, `<>`, or
    `=`, and having that translated directly into a SQL operator
- `facets` (alt: `facet`, `f`)
  - Values supplied to `facets` automatically get wrapped in `%`
    characters, due to the internal representation. You're free to add
    them where you like in values belonging to other attributes

You may have noticed that `//` and `\\` _also_ map to `OR` and
`AND`. You may have also noticed that they only occur within attribute
values. That's because they're syntactic sugar to compactly specify
multiple values for a single attribute:

- `a:x//y` (compact) is equivalent to `a:x || a:y` (expanded)
- This means that more complex, ordered conditions can be constructed
  by using the expanded form, combined with `((` and `))` as needed
  - `((` and `))` are not supported in the compact form, and will lead
    to parse failures or unexpected results

The filter SQL itself uses placeholders. The values from the format
string are held in `c.FltrVals`, and are used in subsequent queries
until a new filter is parsed. The set from the example string is:

`["%funk%", "197%", "1995", "snarky puppy", "confunktion"]`

You can see the `facets` value is wrapped in `%`s, and the specified
`%` in the first `year` value left alone, as described earlier.

The final result of calling `c.ParseFormat` is that `c.FltrCount` will
be set to the count of tracks which match the filter expression.


## Querying

Once a filter 

## tmctool

A CLI utility, `tmctool`, is also provided. It provides basic catalog
maintenance functions:

- Database creation
- Music collection scanning (catalog import + update)
- Applying and removing facets to catalogued tracks
- ID3 tag editing of files (because your music collection is the
  source of truth for metadata other than non-genre facets)


# thrasher-music-catalog

The catalog is the component of the [Thrasher music
service](https://github.com/firepear/thrasher-music-service) suite
which provides faceted classification (i.e. "tagging"), querying, and
metadata management for the music collection.

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

## Config file

To make things easier, all components of Thrasher use a common config
file. The default location is `~/.config/tmcrc` and its format is

```
{
  "dbfile": "/path/to/thrashermusic.db",
  "musicdir": "/path/to/music/files",
  "artist_cutoff": INT
}
```

Of these three, `artist_cutoff` probably needs explanation: it is the
number of tracks that an artist should have in the collection to be
included in the "Artists" listing of the player UI (default: 4).


## Instantiate a catalog instance

`tmc.New` takes two arguments: the path to the on-disk SQLite DB,
and the name to be used for the in-memory copy (which is the working
datastore for the catalog). It returns a `*tmc.Catalog`

```
import (
    tmc "github.com/firepear/thrasher-music-catalog"
)

func main() {
    c, err := tmc.New("/path/to/onDisk.db", "memDbName")
    if err != nil { // as appropriate... }
    // c is ready to use
)
```

## Filtering

The catalog is accessed by first setting a _filter_ and then fetching
tracks. The filter is set by calling `c.Filter` with a _format
string_ argument. An example:

`c.Filter("f:funk&&((y:197*//>=1995))||a:snarky puppy\\confunktion")`

That looks horrible, but the first thing to note is that whitespace is
only significant within attribute values (which we'll come to in a
moment). The second thing to note is that attributes themselves have
expanded forms. The format string can also be written as follows and remain valid:

`facet: funk && (( year: 197% // >=1995 )) || artist: snarky puppy \\ confunktion`

This is a great deal more readable. In fact it resembles the `WHERE`
clause of a SQL query, because that's exactly what it becomes. If we
examine `c.FltrStr` after calling `ParseFormat`, in the middle of it
will be:

`WHERE facets LIKE ? AND ( year LIKE ? OR year >= ? ) OR artist LIKE ? AND artist LIKE ?`

(Not the most sensical query, but a good example of a format string.)

So `&&` and `||` are the logical operators they look like, mapping to
`AND` and `OR`. Doubled parens (`((` and `))`) are escapes for a
single paren in the generated SQL, and are grouping for order of
operations, as expected.

Most of what's left is `attribute: value` pairs, which work exactly
the way you expect them to from JSON or pretty much any other
language, except that no quoting or commas are needed. The supported
attributes are:

- `artist`; short: `a`
- `album`; short: `b`
- `facets`; alt: `facet`, `f`
- `num` (track number); short: `n`
- `title`; short: `t`
- `year`; short: `y`

There are some considerations for attribute values:

- Values, excepting those belonging to `facets` attributes, may be
  prefixed with the standard comparison operators `<=`, `>=`, `<>`,
  and `=`
  - If present, these will be used in the generated SQL
  - As in the example, where `>=1996` became `year >= ?`
- Attribute value wildcards/globs are supported
  - The standard `*` character is used
  - As in the example, `y:197*`
  - Using wildcards with comparison operators will likely give poor
    results

You may have noticed that `//` and `\\` _also_ map to `OR` and
`AND`. This is only valid within attribute values, because it's
syntactic sugar to compactly specify multiple values for a single
attribute:

- `a:x//y` (compact) is equivalent to `a:x || a:y` (expanded)
- This means that more complex, ordered conditions can be constructed
  by using the expanded form, combined with `((` and `))` as needed
  - `((` and `))` are not supported in the compact form, and will lead
    to parse failures or unexpected results

The filter SQL itself uses placeholders. The values from the format
string are held in `c.FltrVals`, and are used in subsequent queries
until a new filter is parsed. The set from the example string is:

`["funk", "197%", "1995", "snarky puppy", "confunktion"]`

The final result of calling `c.ParseFormat` is that `c.FltrCount` will
be set to the number of tracks which match the filter expression.


## Querying

Once a filter has been parsed and set, `c.Query` can be called to
return the paths to the tracks in the filtered set. `Query` takes two
arguments, a limit and an offset.

If you want the entire set, call `Query` with a limit equal to the
size of the filtered set, and an offset of zero:

`trks, err := c.Query(c.FltrCount, 0)`

If you want to paginate the set, provide values for limit and offset
which are appropriate for your application.

### Getting track info

`c.Query` returns a list of paths. To get the remaining data for a
track. call `c.TrkInfo`, which returns an instance of struct
`*tmc.Track`:

```
trks, _ := c.Query(limit, offset)
for _, path := range trks {
    trk := c.TrkInfo(path)
    ...
}
```

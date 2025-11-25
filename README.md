# thrasher-music-catalog

The Catalog is the component of the [Thrasher Music
Service](https://github.com/firepear/thrasher-music-service) suite
which provides faceted classification (i.e. "tagging"), querying, and
metadata management for the music collection.

Unless you are planning to develop a new application which uses this
library, you're probably not interested in this repo. You probably
want either the service (linked earlier) or the
[Tool](https://github.com/firepear/thrasher-music-service).


## Catalog docs

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

### Config file

To make things easier, all components of Thrasher use a common config
file. It is documented under the Tool.

### Instantiate a catalog instance

```
conf, err = tmc.ReadConfig()
cat, err = tmc.New(conf, "DBNAME")
if err != nil {
        fmt.Printf("error creating catalog: %s", err)
        os.Exit(1)
}
defer cat.Close()
cat.TrimPrefix = VALUE
```

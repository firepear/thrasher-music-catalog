# Release notes for thrasher-music-catalog

## v0.9.0 (2026-xx-xx)

- `updater`
  - Database migrations now supported
    - Migration 1: support for future migrations; normalizing `ctime`
      for existing albums


## v0.8.5 (2026-01-28)

- New values added to config
  - `Clientdir` specifies the location of the `thrasher-music-service`
    backend HTTP root directory
  - `TLS` controls whether redirect URLs use `https` or `http`
  - `TLSHost` sets the hostname for redirect URLs
  - `TTL` sets the spawned server time-to-live since last ping recieved


## v0.8.4 (2026-01-27)

- Config file now has a primary location (`/etc/tmc.json`) to support
  containerized ops. The original location (`~/.comfig/tmcrc`) is now
  checked second
- Restore funcs are now in `restore.go`

## v0.8.3 (2026-01-08)

- `c.TrkInfo` was not populating the Ctime or Mtime fields, and no part
  of the system had ever noticed. Now it does
- `c.TrkInfo` now has a second boolean argument, as `c.TrkExists`
  does and serving the same purpose


## v0.8.2 (2025-11-25)

- Config fixes


## v0.8.1 (2025-11-25)

- Added `Listen` field to `tmc.Config` to support service instances
  runing behind proxies


## v0.8.0 (2025-11-25)

- `catalog`
  - `getFacets` now returns a sorted list
  - Fix for tracks being duplicated in queue when multiple matching
    facets have been selected
  - Recent tracks listing now sorted by ctime, tnum
  - Old documentation removed
- `updater`
  - Added `SetAlbum`


## v0.7.0 (2025-07-05)

- `catalog`
  - Fixes to `TrkExists`
- `updater`
  - Added `AddFacet`
  - Added `SetYear`


## v0.6.0 (2025-06-08)

- On scan, APIC data is now extracted in directories which do not
  contain `cover.jpg`
- In 0.5.0, ArtistCutoff was added to the config file, but the value
  was not used. This release corrects that error.
- `tmctool` and all ID3 code removed from Catalog package


## v0.5.0 (2025-05-31)

- Multiple `tmctool` scan/metadata updates
  - Any/all whitespace trimmed from ID3 tag data
  - Track creation time now set to oldest time in `os.Stat` call
  - Track year now set to 9999 if not in ID3 tags
  - Track number now set to 99 if not in ID3 tags
- QueryRecent implemented
- Config struct now exists and is passed to New
  - Config file parsing is now in tmc.ReadConfig; for DRY
- Title and album sorts are now case-insensitive


## v0.4.2 (2025-05-14)

- Add TrackCount to Catalog


## v0.4.1 (2025-05-10)

- Rewrite of chunking implementation for filter interpretation
- Additional fixes for filter chunking


## v0.3.0 (2025-05-09)

- Fixes for filter chunking


## v0.2.0 (2025-05-08)

- Fixes for TrkInfo
- Fixes for facets in queries


## v0.1.0 (2025-05-07)

- Initial release

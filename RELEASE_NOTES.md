# Release notes for thrasher-music-catalog

## v0.9.0 (2025-11-xx)

- `updater`
  - Database migrations now supported
    - Migration 1: support for future migrations; normalizing `ctime`
      for existing albums


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

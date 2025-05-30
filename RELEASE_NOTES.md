# Release notes

## v0.5.0 (2025-05-25)

- Multiple `tmctool` scan/metadata updates
  - Any/all whitespace trimmed from ID3 tag data
  - Track creation time now set to oldest time in `os.Stat` call
  - Track year now set to 9999 if not in ID3 tags
  - Track number now set to 99 if not in ID3 tags
- QueryRecent implemented; returns ~200 most recently modified tracks
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

## 0.3.0 - 2025-02-12

### Fixed

- Fixed links and additional information extraction from noblego.

### Changed

- **[BREAKING]** Changed the video URL type from map to slice.

## 0.2.0 - 2025-02-11

### Added

- Added the `Seek` method to the `fs` storage implementation to search for records using cigar name.
- Added freetext description extraction to the package `extract/nobelgo`.
- Added youtube link extraction to the package `extract/nobelgo`.

## 0.1.0 - 2025-02-10

### Added

- Added the package `extract/nobelgo` to extract data from nobelgo.de.
- Added the command line configuration:
  - output directory;
  - number of records per page to limit data fetching;
  - the page to start data fetching;
  - the page to end data fetching.

## 0.4.1 - 2025-02-15

### Added

- Added the dynamic backoff calculation to the http client.

### Fixed

- Fixed the floating point rounding when storing data after parsing string to float64 and converting cm to mm.
- Fixed extraction of the following attributes from noblego.de:
  - `WrapperTobaccoVariety`;
  - `Format`;
  - `Construction`.

## 0.4.0 - 2025-02-15

### Added

- Added package `extract/cigarworld` to extract data from cigarworld.de.
- Added data mini-batching to persist data between sink of the data from a single page. 

## 0.3.0 - 2025-02-12

### Added

- Added filter to fetch the data from noblego.de corresponding to the cigars with defined gauge.

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

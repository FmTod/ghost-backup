# Changelog

## [0.4.3](https://github.com/FmTod/ghost-backup/compare/v0.4.2...v0.4.3) (2026-01-06)


### Bug Fixes

* **nix:** add bash and coreutils to ghost-backup service path ([a0fe283](https://github.com/FmTod/ghost-backup/commit/a0fe283643fa7cf68ed6bb085d7de0e2665b9175))

## [0.4.2](https://github.com/FmTod/ghost-backup/compare/v0.4.1...v0.4.2) (2026-01-06)


### Bug Fixes

* **nix:** add missing path configuration for ghost-backup service ([b9d46f8](https://github.com/FmTod/ghost-backup/commit/b9d46f8c2728e4ac0cbc91b067df481e191dd4f6))
* **nix:** update nixpkgs locked version and hash in flake.lock ([5cdd354](https://github.com/FmTod/ghost-backup/commit/5cdd3544b7555cb18fe524182556bd1ba9a0a491))

## [0.4.1](https://github.com/FmTod/ghost-backup/compare/v0.4.0...v0.4.1) (2026-01-05)


### Miscellaneous Chores

* release 0.4.1 ([e3f40aa](https://github.com/FmTod/ghost-backup/commit/e3f40aa0d578778ab838822d417c9ce2a81b414c))

## [0.4.0](https://github.com/FmTod/ghost-backup/compare/v0.3.0...v0.4.0) (2026-01-05)


### Features

* add 'only_staged' option to configuration for backing up only staged changes ([51d5a42](https://github.com/FmTod/ghost-backup/commit/51d5a423de1bb9e924079659f0ee9f574cfc64ce))
* add JSON schema support to local configuration ([cc1ff93](https://github.com/FmTod/ghost-backup/commit/cc1ff935073657100d7c9ae71736474aac94b520))


### Bug Fixes

* enhance git repository validation with error handling ([eba6409](https://github.com/FmTod/ghost-backup/commit/eba6409b2d871ef262283abfcc0c98b7cb86f7a0))

## [0.3.0](https://github.com/FmTod/ghost-backup/compare/v0.2.0...v0.3.0) (2026-01-05)


### Features

* configure release-please to update package.json and nix/package.nix versions ([#9](https://github.com/FmTod/ghost-backup/issues/9)) ([4f639dd](https://github.com/FmTod/ghost-backup/commit/4f639dd4d25909a93bab31280beb21e2b901792a))


### Bug Fixes

* update conditions for triggering Nix package update workflow ([56de38f](https://github.com/FmTod/ghost-backup/commit/56de38ff7de6c0fd6e5a747dffd48e4a3ac1b244))

## 0.2.0 (2026-01-05)


### Features

* add workflow to update nix/package.nix on release ([5f2c3a9](https://github.com/FmTod/ghost-backup/commit/5f2c3a922a84d4ea9d5947a77af9b2312ed1ae99))


### Bug Fixes

* add robust fallback patterns for vendor hash extraction ([7a0996e](https://github.com/FmTod/ghost-backup/commit/7a0996e94a2dd50924c64b9e1cfed834bf0f357e))
* clean up workflow formatting and remove trailing spaces ([a3a5da1](https://github.com/FmTod/ghost-backup/commit/a3a5da10d01fd3241fda1e7fd097f815ad462dfa))


### Miscellaneous Chores

* release 0.2.0 ([ccf729e](https://github.com/FmTod/ghost-backup/commit/ccf729ec52eef4c3712980593014c6ff575f46b3))

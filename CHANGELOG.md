# Changelog

## [0.1.8](https://github.com/hostinger/fireactions/compare/v0.1.7...v0.1.8) (2024-03-18)


### Bug Fixes

* Don't set expiration time on Containerd lease ([78e76f1](https://github.com/hostinger/fireactions/commit/78e76f18571622e405f93a990526155f099dbf2b))

## [0.1.7](https://github.com/hostinger/fireactions/compare/v0.1.6...v0.1.7) (2024-03-16)


### Bug Fixes

* Race condition on graceful shutdown ([5ad2390](https://github.com/hostinger/fireactions/commit/5ad23903740cf1b54645cb97bd40f7ab83c74c72))

## [0.1.6](https://github.com/hostinger/fireactions/compare/v0.1.5...v0.1.6) (2024-03-16)


### Bug Fixes

* Deadlock when removing Firecracker VM reference from map ([ee5c00a](https://github.com/hostinger/fireactions/commit/ee5c00ac61df9065709b51e14b0113d3c0925c0f))
* Gracefully handle SIGTERM ([0302167](https://github.com/hostinger/fireactions/commit/0302167b3c4cd34fe1c9fa1ae8202697d4ef42c4))

## [0.1.5](https://github.com/hostinger/fireactions/compare/v0.1.4...v0.1.5) (2024-03-15)


### Bug Fixes

* Don't set timeout on Firecracker VM context ([3570347](https://github.com/hostinger/fireactions/commit/3570347149bb99348a345f6e4fb3b55301ef8907))

## [0.1.4](https://github.com/hostinger/fireactions/compare/v0.1.3...v0.1.4) (2024-03-15)


### Bug Fixes

* Collect current_runners_count metric even if pool is paused ([0c0969c](https://github.com/hostinger/fireactions/commit/0c0969c25a9696bd904c617419ad2cb8aeef1247))

## [0.1.3](https://github.com/hostinger/fireactions/compare/v0.1.2...v0.1.3) (2024-03-15)


### Bug Fixes

* Correctly set fireactions_pool_scale_ metrics ([9a6988b](https://github.com/hostinger/fireactions/commit/9a6988b9452cd676a3b082e94213d4b0321d9e69))

## [0.1.2](https://github.com/hostinger/fireactions/compare/v0.1.1...v0.1.2) (2024-03-13)


### Bug Fixes

* **goreleaser:** Correct checksum name format ([463008e](https://github.com/hostinger/fireactions/commit/463008ef27dd0a1951dfdc6eafa4d772aac20ea5))

## [0.1.1](https://github.com/hostinger/fireactions/compare/v0.1.0...v0.1.1) (2024-03-13)


### Bug Fixes

* **goreleaser:** Include v prefix in package name ([4be3e03](https://github.com/hostinger/fireactions/commit/4be3e033b563785a53252f1c8ac23d5b9925597f))

## [0.1.0](https://github.com/hostinger/fireactions/compare/v0.0.1...v0.1.0) (2024-03-13)


### Features

* Initial commit ([b996018](https://github.com/hostinger/fireactions/commit/b9960186c7eb695fbb0a8c59f8194d8604e72ee4))

# Changelog

## [2.0.0](https://github.com/hostinger/fireactions/compare/v1.3.2...v2.0.0) (2023-11-28)


### ⚠ BREAKING CHANGES

* Simplify configuration for both server and client
* Multiple refactorings

### Features

* **client:** Add optional CLI flags ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))
* Reconcile assigned runners with adjustable concurrency. ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))
* Simplify configuration for both server and client ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))
* Use a shorter string ID instead of UUID ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))
* Use cbrgm/githubevents instead of go-playground/webhooks ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))
* Use Just-In-Time(JIT) runner configuration instead of registration/removal tokens. ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))


### Code Refactoring

* Multiple refactorings ([8a19dd9](https://github.com/hostinger/fireactions/commit/8a19dd9647c997e3ab5d345c156838f4048201ae))

## [1.3.2](https://github.com/hostinger/fireactions/compare/v1.3.1...v1.3.2) (2023-11-22)


### Bug Fixes

* **server:** Deregistering a node should remove assigned runners as well ([016adff](https://github.com/hostinger/fireactions/commit/016adffae55a68ef10de00e72d0867260eb03f09))

## [1.3.1](https://github.com/hostinger/fireactions/compare/v1.3.0...v1.3.1) (2023-11-21)


### Bug Fixes

* Bind env variables with BindEnv() instead of AutomaticEnv() due to viper bug ([19da70b](https://github.com/hostinger/fireactions/commit/19da70bd2ccc4d4745231aba75390dd2c2c52f3f))

## [1.3.0](https://github.com/hostinger/fireactions/compare/v1.2.0...v1.3.0) (2023-11-19)


### Features

* Ability to set metadata for MMDS ([948ef5b](https://github.com/hostinger/fireactions/commit/948ef5b912a1af460b373056efcdd462f3ac9de2))
* **images:** Set default /etc/resolv.conf ([0686427](https://github.com/hostinger/fireactions/commit/06864276a9641773c4f33594c9b569b035fa3c3a))
* **images:** Set metadata.fireactions.internal alias for MMDS ([fa1e4e8](https://github.com/hostinger/fireactions/commit/fa1e4e8ef2ec9a56ea103f8308ff7ff31f161e7c))
* Rearchitecture fireactions-agent into HTTP service ([0c3809c](https://github.com/hostinger/fireactions/commit/0c3809cbb7c9193ca64ea23f0443af8259051aee))


### Bug Fixes

* **server:** Add missing JSON tags ([479b4d5](https://github.com/hostinger/fireactions/commit/479b4d540634e3906f24f711e24410308845989d))
* **server:** Validate Config.GitHub.JobLabel.AllowedRepositories regexp ([1176e92](https://github.com/hostinger/fireactions/commit/1176e92f30c8ff1062ecb005333bef18512b05b0))

## [1.2.0](https://github.com/hostinger/fireactions/compare/v1.1.0...v1.2.0) (2023-11-08)


### Features

* **ci:** Build multiplatform Docker image ([4ce43b3](https://github.com/hostinger/fireactions/commit/4ce43b3844cfe4483e6a0b9dc4ab2ca04b72810e))

## [1.1.0](https://github.com/hostinger/fireactions/compare/v1.0.3...v1.1.0) (2023-11-08)


### Features

* **cli:** Add --columns flag to runners list command ([afd9568](https://github.com/hostinger/fireactions/commit/afd95681f0e3627157a0590a041fba762fbd91b1))
* **cli:** Add runners create command ([e68a91b](https://github.com/hostinger/fireactions/commit/e68a91b87e145f22f649909cd848091522d5ec0e))
* **cli:** Replace version command with --version flag ([da54c12](https://github.com/hostinger/fireactions/commit/da54c12aba9c3e7469a31905f48639ab18f6a04a))


### Bug Fixes

* **agent:** Remove duplicate --replace flag ([afd8b19](https://github.com/hostinger/fireactions/commit/afd8b196457186251eac76e14638e3a6f1fc9aee))
* **client:** Close containerd client from runner.Manager ([70c2c77](https://github.com/hostinger/fireactions/commit/70c2c771efea58a514b9d3928299f0593ab38d25))
* **client:** Ensure runners in Completed status are stopped ([f282015](https://github.com/hostinger/fireactions/commit/f282015a5334b35ab8cd8b52be33e68994c7bcdd))
* **client:** Fix concurrent map writes ([b094577](https://github.com/hostinger/fireactions/commit/b094577b9b6a5b5ad29cef2ffaae7825b99b7bbe))
* **server:** Change default JobLabelPrefix fireactions -&gt; fireactions- ([7320308](https://github.com/hostinger/fireactions/commit/7320308e8db5705cfacc692bb3a866df9ce4b1ec))
* **server:** Don't requeue deleted runners to scheduling queue ([5c1e0e5](https://github.com/hostinger/fireactions/commit/5c1e0e58ebb1de9ff37c988954990b212486e520))
* **server:** Set runner name prefix to fireactions- ([0544c09](https://github.com/hostinger/fireactions/commit/0544c09e2b7b883e8047498a8faf144d7e11f315))

## [1.0.3](https://github.com/hostinger/fireactions/compare/v1.0.2...v1.0.3) (2023-10-31)


### Bug Fixes

* **ci:** Link GHCR package to repository ([83825b4](https://github.com/hostinger/fireactions/commit/83825b4d3fcccd7e8f8760dc4b084255cbc1047c))

## [1.0.2](https://github.com/hostinger/fireactions/compare/v1.0.1...v1.0.2) (2023-10-31)


### Bug Fixes

* **ci:** Set missing step ID on release workflow ([30cbf01](https://github.com/hostinger/fireactions/commit/30cbf0120762fb50779beded95c6bb7950e720ff))

## [1.0.1](https://github.com/hostinger/fireactions/compare/v1.0.0...v1.0.1) (2023-10-31)


### Bug Fixes

* **ci:** Set missing release_created output ([1e9cc1f](https://github.com/hostinger/fireactions/commit/1e9cc1f736afba9c86c6aa720b3b0bde8b1b4ad0))

## 1.0.0 (2023-10-31)


### Features

* Initial commit ([9fd316a](https://github.com/hostinger/fireactions/commit/9fd316a3b341860506aa86ffceda50a6703963f4))


### Miscellaneous Chores

* Release v1.0.0 ([dad20ed](https://github.com/hostinger/fireactions/commit/dad20ed3f2a275c624fc6a0bd4625d536abed7cb))

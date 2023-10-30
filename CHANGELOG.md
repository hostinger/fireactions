# Changelog

## [1.3.0](https://github.com/hostinger/fireactions/compare/v1.2.0...v1.3.0) (2023-10-30)


### Features

* Add /healthz endpoint ([a7742c8](https://github.com/hostinger/fireactions/commit/a7742c84b30c5bb631ab686b049745fc21cafb25))
* Add /livez, /readyz, /version handlers ([d958104](https://github.com/hostinger/fireactions/commit/d95810441c3ddfed5e418a85a8810cf84dee33ac))
* Add client ImageSyncer and ImageGC implementations ([b58bd93](https://github.com/hostinger/fireactions/commit/b58bd936d07ee3be325d59b3dc2acaf587fa3ab1))
* Add HeartbeatFilter to scheduler ([144a18c](https://github.com/hostinger/fireactions/commit/144a18c59f7a079c7fe875189dec6be926fa8c81))
* Add initial preflight checks for client ([80123ae](https://github.com/hostinger/fireactions/commit/80123aec2ce0f05b2d7bbdba8113741a67e3eafa))
* Add initial Prometheus metrics ([e16a124](https://github.com/hostinger/fireactions/commit/e16a124f7827611e5a74a2ea33605a8ada915b58))
* Add nodes deregister command ([6164547](https://github.com/hostinger/fireactions/commit/6164547fd8d37d7f95f9bc6b0186212711053235))
* Add preflight command ([6f4a543](https://github.com/hostinger/fireactions/commit/6f4a543c00f5582ff76348c72905cecd32e2f4a6))
* Add version command ([293746d](https://github.com/hostinger/fireactions/commit/293746db4b352f12a264e4f7684e8a7cf502f237))
* Allow disabling/enabling flavors ([d2d077c](https://github.com/hostinger/fireactions/commit/d2d077cb7d231ea6c871da02a5b97c54cf24b70f))
* Allow disabling/enabling groups ([9341404](https://github.com/hostinger/fireactions/commit/93414046272d12ad7a0e379730c32e524a6477b1))
* Allow to cordon/uncordon nodes from scheduling ([1001aa9](https://github.com/hostinger/fireactions/commit/1001aa9665923075eff0bfc5e67d36530d78a33e))
* Client can belong to many groups ([9fffcd6](https://github.com/hostinger/fireactions/commit/9fffcd6c492b6e3c081a1c6777b4268982e929c9))
* Display CPU,MEM usage in Node printer ([f86eb7f](https://github.com/hostinger/fireactions/commit/f86eb7f8e1341a97e836852a080c71eb44ab21ae))
* Endpoint DELETE /api/v1/flavors/:name ([acc9fff](https://github.com/hostinger/fireactions/commit/acc9fff39dec164d0a8ac764b87b940f4d15e58f))
* Endpoint DELETE /api/v1/groups/:name ([ce3fcbc](https://github.com/hostinger/fireactions/commit/ce3fcbc8fa74e5e75ac644fb94c1855c9337a5b5))
* Endpoint POST /api/v1/groups ([80d54a3](https://github.com/hostinger/fireactions/commit/80d54a33d7e8468d3e9fbf793102b50697e156ea))
* Endpoint POST /api/v1/images ([5560614](https://github.com/hostinger/fireactions/commit/55606144849016ff4620f46fdd2957adb5393ca7))
* Flavors of GitHub runners ([3473769](https://github.com/hostinger/fireactions/commit/3473769f207e05bdafa7eb96383bb3c61f98cb3a))
* Images API ([edf53fb](https://github.com/hostinger/fireactions/commit/edf53fb134bdd8d51a98840e75d9c4f50196204d))
* Implement groups for grouping node(s) to runner(s) ([3a32c64](https://github.com/hostinger/fireactions/commit/3a32c642c0aaa02fef6fd0329c35abe68fc2a1c4))
* Initial commit ([9c98f55](https://github.com/hostinger/fireactions/commit/9c98f55c8a8f1b0c0056cc6ce48361cba9f130cd))
* Set default flavor and group dynamically ([f49b1bd](https://github.com/hostinger/fireactions/commit/f49b1bdb52428b4d27f014cb93d57cc83d423344))
* Store client registration info locally ([389c376](https://github.com/hostinger/fireactions/commit/389c3761cb95c744e56630f1d488931fe4192148))
* The Great refactoring ([b7edac2](https://github.com/hostinger/fireactions/commit/b7edac270732d13a1975fcb8a12ca8d934899368))
* Use '.' instead of '-' as job label separator ([d6f7bd3](https://github.com/hostinger/fireactions/commit/d6f7bd3dcfb797e4e267e2fd6ca97cd839311bb0))
* Use underscores in configuration ([c9c67af](https://github.com/hostinger/fireactions/commit/c9c67affd13715f632eae3f2124f3a22a3b7c378))


### Bug Fixes

* Add config search paths for client ([c1ba0f4](https://github.com/hostinger/fireactions/commit/c1ba0f43e97877e4477060a36e8441434c40a96e))
* Add node to scheduler's internal cache on registration ([98f6262](https://github.com/hostinger/fireactions/commit/98f6262165e187ff424d0628451a923617bcc701))
* Convert structs.Node to v1.Node only if it's not nil ([5a8b3ba](https://github.com/hostinger/fireactions/commit/5a8b3ba461076c5e4aefd807fbe572152f4ed30f))
* Deallocate node resources only on Delete call ([79dc599](https://github.com/hostinger/fireactions/commit/79dc599123ad28674f531f52b39b65e69c13b6df))
* Don't reschedule completed runners on start ([be72ca9](https://github.com/hostinger/fireactions/commit/be72ca92a18405d19301464d5bb6b994f0d1d1bc))
* Make scheduler configuration optional ([076370a](https://github.com/hostinger/fireactions/commit/076370a0737b8686d05cca87db43a2ef5b99f196))
* Remove unused DefaultJobLabel config option ([8337ac5](https://github.com/hostinger/fireactions/commit/8337ac5f088950019b55ac3ab1aca332bfded65a))
* Server environment variable parsing ([481f46b](https://github.com/hostinger/fireactions/commit/481f46b024c37bbac0265f148692ae4535bc50df))

## [1.2.0](https://github.com/hostinger/fireactions/compare/v1.1.0...v1.2.0) (2023-10-30)


### Features

* Add /livez, /readyz, /version handlers ([d958104](https://github.com/hostinger/fireactions/commit/d95810441c3ddfed5e418a85a8810cf84dee33ac))


### Bug Fixes

* Server environment variable parsing ([481f46b](https://github.com/hostinger/fireactions/commit/481f46b024c37bbac0265f148692ae4535bc50df))

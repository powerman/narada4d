# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.5] - 2026-07-12

[1.8.5]: https://github.com/powerman/narada4d/compare/v1.8.4..v1.8.5

## [1.8.4] - 2026-07-12

[1.8.4]: https://github.com/powerman/narada4d/compare/v1.8.3..v1.8.4

## [1.8.3] - 2026-06-10

### 🐛 Fixed

- **(mise)** Add dockerize by @powerman in [a979dd8]
- Make goose-postgres & goose-mysql locks immune to caller timeouts by @powerman in [ddacd53]

[1.8.3]: https://github.com/powerman/narada4d/compare/v1.8.2..v1.8.3
[ddacd53]: https://github.com/powerman/narada4d/commit/ddacd53d4751bca1ebd2c7bb90bdd3778df1ff37
[a979dd8]: https://github.com/powerman/narada4d/commit/a979dd897cd1962064a9195b3cb011166636c323

## [1.8.2] - 2026-06-10

[1.8.2]: https://github.com/powerman/narada4d/compare/v1.8.1..v1.8.2

## [1.8.1] - 2026-06-10

### 🐛 Fixed

- **(release)** Build both binaries for all platforms by @powerman in [29d7805]

### 📚 Documentation

- Note that file protocol is not supported on Windows by @powerman in [06e9cc9]

[1.8.1]: https://github.com/powerman/narada4d/compare/v1.8.0..v1.8.1
[06e9cc9]: https://github.com/powerman/narada4d/commit/06e9cc97660ccde6f8dd5924229c349b6517ec64
[29d7805]: https://github.com/powerman/narada4d/commit/29d7805abfecd9d8dc8e4225c039a8a3522fb64f

## [1.8.0] - 2026-06-10

### 🔔 Changed

- Upgrade Go to 1.25 and deps by @powerman in [7ff3283]
- Disable protocol/file on Windows by @powerman in [7c568aa]

### 🐛 Fixed

- **(release)** Build both binaries (narada4d-init and narada4d-lock) by @powerman in [8513c02]

[1.8.0]: https://github.com/powerman/narada4d/compare/v1.7.1..v1.8.0
[7ff3283]: https://github.com/powerman/narada4d/commit/7ff328348a177920890315fb279305ba04cf44f5
[7c568aa]: https://github.com/powerman/narada4d/commit/7c568aa18136e80a7e3ea3d94f3b1b7e85b49f9d
[8513c02]: https://github.com/powerman/narada4d/commit/8513c0205388f24460bbd5a68ff254f709630064

## [1.7.1] - 2020-10-23

### 🐛 Fixed

- Panic on reconnect by @powerman in [#24]

[1.7.1]: https://github.com/powerman/narada4d/compare/v1.7.0..v1.7.1
[#24]: https://github.com/powerman/narada4d/pull/24

## [1.7.0] - 2020-04-18

### 🚀 Added

- Support list of URLs in `NARADA4D_SKIP_LOCK` by @powerman in [36e6dbc]

[1.7.0]: https://github.com/powerman/narada4d/compare/v1.6.1..v1.7.0
[36e6dbc]: https://github.com/powerman/narada4d/commit/36e6dbc09ac7da78108cf67aeaf6d8497d9615f1

## [1.6.1] - 2020-04-03

[1.6.1]: https://github.com/powerman/narada4d/compare/v1.6.0..v1.6.1

## [1.6.0] - 2020-04-02

### 🚀 Added

- Add HoldSharedLock by @powerman in [3da898b]

[1.6.0]: https://github.com/powerman/narada4d/compare/v1.5.0..v1.6.0
[3da898b]: https://github.com/powerman/narada4d/commit/3da898bad84c64e5a047a130ee75fcbb124c2772

## [1.5.0] - 2020-04-01

### 🔔 Changed

- Optimize SharedLock by @powerman in [0cb3f2b]

[1.5.0]: https://github.com/powerman/narada4d/compare/v1.4.0..v1.5.0
[0cb3f2b]: https://github.com/powerman/narada4d/commit/0cb3f2b39ac705b3fa319708e76279909c971574

## [1.4.0] - 2020-04-01

### 🚀 Added

- Add protocol/goose-mysql by @powerman in [e8ac45d]
- Add NewAt by @powerman in [2655013]

### 🔔 Changed

- Auto-initialization in New by @powerman in [af2ecb0]

[1.4.0]: https://github.com/powerman/narada4d/compare/v1.3.0..v1.4.0
[e8ac45d]: https://github.com/powerman/narada4d/commit/e8ac45d97f9d029234d65d505acc2b6e1c101a03
[2655013]: https://github.com/powerman/narada4d/commit/2655013080e31a272fe307dc92d9208ace881cb9
[af2ecb0]: https://github.com/powerman/narada4d/commit/af2ecb0087aee83e0bcab32bc14256be1efb90e8

## [1.3.0] - 2020-02-01

### 🔔 Changed

- Bump github.com/powerman/gotest from 0.2.0 to 0.3.0 by @dependabot-preview[bot] in [#5]

### 🐛 Fixed

- Tests by @powerman in [57563b9]

[1.3.0]: https://github.com/powerman/narada4d/compare/v1.2.1..v1.3.0
[#5]: https://github.com/powerman/narada4d/pull/5
[57563b9]: https://github.com/powerman/narada4d/commit/57563b94ecb3102dcc75aefc03ad8629148b6f10

## [1.2.1] - 2019-07-27

### 🐛 Fixed

- Go get by @powerman in [f855eb9]

[1.2.1]: https://github.com/powerman/narada4d/compare/v1.2.0..v1.2.1
[f855eb9]: https://github.com/powerman/narada4d/commit/f855eb99bf22919bba10448b7fdfdaa888dae053

## [1.2.0] - 2019-07-27

### 🚀 Added

- Add Close by @powerman in [60c9af9]

[1.2.0]: https://github.com/powerman/narada4d/compare/v1.1.0..v1.2.0
[60c9af9]: https://github.com/powerman/narada4d/commit/60c9af98529063616781f5d057487923d978dce7

## [1.1.0] - 2019-07-19

### 🚀 Added

- Add `protocol/goose_postgres` by @powerman in [a55c826]

### 🔔 Changed

- Cleanup by @powerman in [679e04d]

### 🐛 Fixed

- Happy linter, add protocols to tools by @powerman in [55ccd79]

[1.1.0]: https://github.com/powerman/narada4d/compare/v1.0.0..v1.1.0
[679e04d]: https://github.com/powerman/narada4d/commit/679e04d70b7686aff0c7e63d12f4a37fa814977d
[a55c826]: https://github.com/powerman/narada4d/commit/a55c8267de358d88be82d7fb9c22d6984323aaef
[55ccd79]: https://github.com/powerman/narada4d/commit/55ccd79e3767f7ea18801627ccc406c84ff1eb7f

## [1.0.0] - 2018-11-15

### 🔔 Changed

- Initial commit by @powerman in [1e17193]
- Update README by @powerman in [f928973]
- Initial version by @powerman in [014c32a]
- Add mysql protocol by @powerman in [ddf79c9]
- Massive refactoring by @powerman in [265a8b0]
- Protocol/file: check for initialized dir by @powerman in [7e80f74]
- Cleanup by @powerman in [e34b180]
- Add narada4d-init, narada4d-lock by @powerman in [af3f7c4]
- Add test-plan for protocol/file by @powerman in [7127b4b]
- Ready `proto_tests` by @salamandra19 in [4981b0e]
- `test_schemaver` by @salamandra19 in [08683c3]
- `test_plan` by @salamandra19 in [0136b10]
- Add mock for schemaver tests by @powerman in [2b0d356]
- Unlock and Callback panics added by @salamandra19 in [9c16a02]
- Improve tests by @powerman in [bea258b]
- Cleaned, callback-panic test added by @salamandra19 in [ff89d4b]
- Cleanup by @salamandra19 in [4a6e5f1]
- Raw mysql by @salamandra19 in [0af7b26]
- Funcs connect and parse added by @salamandra19 in [f105086]
- Cleanned by @salamandra19 in [2ff424a]
- Mysql ready by @salamandra19 in [4bbe6f3]
- Version check added by @salamandra19 in [c1386df]
- `tests_mysql` raw by @salamandra19 in [826c20c]
- Move to protocol/mysql by @powerman in [86d9557]
- Add docker for tests by @powerman in [66ee6fa]
- Add validate(schema check), Set(fix regexp) by @salamandra19 in [ba6b9b6]
- More tests ready by @salamandra19 in [e7bba81]
- Mysql cleaned by @salamandra19 in [6cb4018]
- Mysql tests completed by @salamandra19 in [c26977d]
- `mysql_tests` ready by @salamandra19 in [c86f830]
- Fix test by @powerman in [eb75911]
- Add transaction by @powerman in [2772e63]
- Cleanup by @powerman in [a730176]
- Use mysql:5.6 by @powerman in [e7ef8a4]
- Fix collation by @powerman in [00c2f9b]
- Add go.mod by @powerman in [3f3b349]
- Setup linter by @powerman in [7b1fb74]
- Add CircleCI by @powerman in [f04c325]

### 🐛 Fixed

- Fix Set() by @salamandra19 in [840f0e4]
- Fix prev commit by @powerman in [c6a347e]

[1.0.0]: https://github.com/powerman/narada4d/compare/%40%7B10year%7D..v1.0.0
[1e17193]: https://github.com/powerman/narada4d/commit/1e17193b0cd7de249018b752d5fd7bf5cbf38edf
[f928973]: https://github.com/powerman/narada4d/commit/f9289737e54456ee9745917e9264618acbd9ab8c
[014c32a]: https://github.com/powerman/narada4d/commit/014c32a5f20c3a2f8e6340b78b18690ed9c967bd
[ddf79c9]: https://github.com/powerman/narada4d/commit/ddf79c989d848dc4e9c0f1a59fd258e078091cfa
[265a8b0]: https://github.com/powerman/narada4d/commit/265a8b0c9bd148a58462bdfffbbd408346be3f41
[7e80f74]: https://github.com/powerman/narada4d/commit/7e80f748d572159ed53bfc5d2b503e6cb8396425
[e34b180]: https://github.com/powerman/narada4d/commit/e34b1803c3fbf51fa52074b701f1948bea3508f7
[af3f7c4]: https://github.com/powerman/narada4d/commit/af3f7c4f371d50966519ac0c331c4b93552a0992
[7127b4b]: https://github.com/powerman/narada4d/commit/7127b4b3395e8997555da31720d6090cfc6480a4
[840f0e4]: https://github.com/powerman/narada4d/commit/840f0e4d06e68a28061aa8cdd010270838867644
[4981b0e]: https://github.com/powerman/narada4d/commit/4981b0eb4d21ea81a42b779c2f1c80736097b864
[08683c3]: https://github.com/powerman/narada4d/commit/08683c32d329521a93e3ee1dd7e797f21a80593b
[0136b10]: https://github.com/powerman/narada4d/commit/0136b10f907cccbc3deb88b52617f14d62086ff2
[2b0d356]: https://github.com/powerman/narada4d/commit/2b0d356246ba71986c5da77be4b968acdc16122e
[9c16a02]: https://github.com/powerman/narada4d/commit/9c16a02739c8ae9e03893231da32ba7640d19912
[bea258b]: https://github.com/powerman/narada4d/commit/bea258b875264119564117adeb6a06653ee933ce
[ff89d4b]: https://github.com/powerman/narada4d/commit/ff89d4b4bdd0aa5e582d05093645f1dc64ccbc08
[4a6e5f1]: https://github.com/powerman/narada4d/commit/4a6e5f1e3d975bfa60010c1578866463e962fecb
[0af7b26]: https://github.com/powerman/narada4d/commit/0af7b26d94ec327707defcd776f17b6ff36964a8
[f105086]: https://github.com/powerman/narada4d/commit/f105086b544403e586c8d4a2315bd1b151d77757
[2ff424a]: https://github.com/powerman/narada4d/commit/2ff424ac6a8c9caa249be2781e599cd40535e716
[4bbe6f3]: https://github.com/powerman/narada4d/commit/4bbe6f3b8140c64a9fc28bd6c48f365f30d9126d
[c1386df]: https://github.com/powerman/narada4d/commit/c1386df3065bc4ff1efd916e7cdffa9e0d707d7e
[826c20c]: https://github.com/powerman/narada4d/commit/826c20ce83e8be6bd8b3a788ad7604247d6fb2ff
[86d9557]: https://github.com/powerman/narada4d/commit/86d95574c45f7be7b25cc8cf1411b6354bd5c57c
[66ee6fa]: https://github.com/powerman/narada4d/commit/66ee6fa5debc35084dc6321f8819179f4eac6727
[ba6b9b6]: https://github.com/powerman/narada4d/commit/ba6b9b6b57273c66b77bc254a6aae1c8a7f16cf5
[e7bba81]: https://github.com/powerman/narada4d/commit/e7bba810602e15b7b9dd089e5566ceccdf09ae9d
[6cb4018]: https://github.com/powerman/narada4d/commit/6cb40188ad2a5eb2fe30e6b94f3cc5e83a46cb03
[c26977d]: https://github.com/powerman/narada4d/commit/c26977dd624c75fb91d6ef3808979f8e62a9ca4e
[c86f830]: https://github.com/powerman/narada4d/commit/c86f83006298d52467c418aceed1b37f05d9207e
[eb75911]: https://github.com/powerman/narada4d/commit/eb7591153cfda8f2e8871af2b9dc036f1f179dbf
[2772e63]: https://github.com/powerman/narada4d/commit/2772e63200f0502dd63823084db9f27d3023b26d
[a730176]: https://github.com/powerman/narada4d/commit/a73017675298bf847c841b76d04eb0abdd6f4079
[e7ef8a4]: https://github.com/powerman/narada4d/commit/e7ef8a475dd990a44cec69821697a263e4a01032
[00c2f9b]: https://github.com/powerman/narada4d/commit/00c2f9bf8eecec8ce0704f4e8728a7f29f238824
[c6a347e]: https://github.com/powerman/narada4d/commit/c6a347e4815b7c8961a4374617b1e5eb53f12342
[3f3b349]: https://github.com/powerman/narada4d/commit/3f3b3496a6b3d24e8b8be26f68905bc563790040
[7b1fb74]: https://github.com/powerman/narada4d/commit/7b1fb74f8f8dae1facfb07c064bd5612bb0495fa
[f04c325]: https://github.com/powerman/narada4d/commit/f04c3259511d55b39bda1ba2bc1bfd2045df760c

<!-- generated by git-cliff -->

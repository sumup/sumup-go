# Changelog

## [0.8.0](https://github.com/sumup/sumup-go/compare/v0.7.0...v0.8.0) (2025-12-01)


### Features

* handle properly initialisms for well-known strings ([#144](https://github.com/sumup/sumup-go/issues/144)) ([ee7b270](https://github.com/sumup/sumup-go/commit/ee7b270e203f7fdccbb714a1e3e1c3f70984b953))

## [0.7.0](https://github.com/sumup/sumup-go/compare/v0.6.0...v0.7.0) (2025-11-30)


### Features

* **example:** oauth2 flow ([#130](https://github.com/sumup/sumup-go/issues/130)) ([be3b6f1](https://github.com/sumup/sumup-go/commit/be3b6f194fb99eca985b09efa967885f57a01c68))


### Bug Fixes

* **ci:** use bot for release PRs ([72bdf15](https://github.com/sumup/sumup-go/commit/72bdf155b3450c163350ce53a3acc1b1e33e2811))

## [0.6.0](https://github.com/sumup/sumup-go/compare/v0.5.0...v0.6.0) (2025-11-11)


### Features

* release 0.6.0 ([3aab4be](https://github.com/sumup/sumup-go/commit/3aab4be626018da1b53eba7671a6674621ac70ca))

## [0.5.0](https://github.com/sumup/sumup-go/compare/v0.4.0...v0.5.0) (2025-11-04)


### Features

* **docs:** update README title ([f874ad2](https://github.com/sumup/sumup-go/commit/f874ad2c1d2bca3b5cf11f16c6a4e7cb6a3f7e59))
* improved typing ([#123](https://github.com/sumup/sumup-go/issues/123)) ([4d80c5e](https://github.com/sumup/sumup-go/commit/4d80c5ee51718866a28374cbf11d2230a431b0ef))

## [0.4.0](https://github.com/sumup/sumup-go/compare/v0.3.0...v0.4.0) (2025-11-03)


### Features

* **example:** add full example of checkout ([#115](https://github.com/sumup/sumup-go/issues/115)) ([e4f9796](https://github.com/sumup/sumup-go/commit/e4f9796ac60ab240bd5dd9e128087bc29eb29260))


### Bug Fixes

* **cd:** commit generated SDK using SumUp Bot ([31b58f1](https://github.com/sumup/sumup-go/commit/31b58f192d4fb93e6425363e134e3866559d5729))
* user-agent string ([7f6fd8a](https://github.com/sumup/sumup-go/commit/7f6fd8aa72611a0c043a95082ed47dcdc0ebcf17))

## [0.3.0](https://github.com/sumup/sumup-go/compare/v0.2.0...v0.3.0) (2025-10-23)

0.3.0 bring the Merchants API, allowing access to multiple merchant accounts, depending on the authorization. For users that authenticate using SumUp's SSO you can now access any of the merchant accounts that they have membership in. For API keys the access is still restricted to the merchant account for which the API key was created. We are working on introducing more authentication options to make integrations that need to rely on multiple merchant accounts easier in the future.

The merchants endpoints replace the legacy `/me/` endpoints and further cleanup the underlying models.

### Features

* update makefile targets ([a21676d](https://github.com/sumup/sumup-go/commit/a21676d1c03295fe5244354aa676b239cf24e8b4))


### Bug Fixes

* **cd:** generated docs repo url ([615e80f](https://github.com/sumup/sumup-go/commit/615e80f29ec39ef794b078679500e30fbd59dd74))

## [0.2.0](https://github.com/sumup/sumup-go/compare/v0.1.0...v0.2.0) (2025-10-02)


### Features

* **ci:** use native govulncheck SARIF functionality ([#103](https://github.com/sumup/sumup-go/issues/103)) ([156150c](https://github.com/sumup/sumup-go/commit/156150c339006e902cbd27afa605a96c3215b12e))

## [0.1.0](https://github.com/sumup/sumup-go/compare/v0.0.1...v0.1.0) (2025-07-12)


### Features

* **readme:** docs badge ([#85](https://github.com/sumup/sumup-go/issues/85)) ([7d2b120](https://github.com/sumup/sumup-go/commit/7d2b120ddaa35ed50a1bd91118bd00c2262d6c6f))

## 0.0.1 (2025-06-17)


### Features

* **cd:** send notification on release ([#32](https://github.com/sumup/sumup-go/issues/32)) ([013a4bb](https://github.com/sumup/sumup-go/commit/013a4bb967730579921e678044fa96fcc8ad48cf))
* **ci/cd:** update actions, add dependabot ([#3](https://github.com/sumup/sumup-go/issues/3)) ([43b02ff](https://github.com/sumup/sumup-go/commit/43b02ff3ba7641e81c2e3bbbf524a8b33b456525))
* **ci:** auto-generate latest SDK ([#5](https://github.com/sumup/sumup-go/issues/5)) ([2343544](https://github.com/sumup/sumup-go/commit/2343544a078da2e6511580048239191ecf008505))
* **ci:** commit generated code ([#6](https://github.com/sumup/sumup-go/issues/6)) ([26cc46b](https://github.com/sumup/sumup-go/commit/26cc46bd644cfd37d81b456ac9babfb680a52d22))
* **ci:** lint github actions ([#22](https://github.com/sumup/sumup-go/issues/22)) ([da1168f](https://github.com/sumup/sumup-go/commit/da1168f1f592af898e3de5ffa992d2e7999b5221))
* doc.go ([44434a7](https://github.com/sumup/sumup-go/commit/44434a7cbf963022c5ddfe7eb85c66069eefb50e))
* **docs:** add link to API reference and developer portal ([#29](https://github.com/sumup/sumup-go/issues/29)) ([ad2507e](https://github.com/sumup/sumup-go/commit/ad2507e4036f058a95b344d268c2a63d9a9f1ecd))
* **docs:** README badges ([28627fb](https://github.com/sumup/sumup-go/commit/28627fb8bebc3bfce09ec39b6b331331112aa18a))
* **docs:** security policy ([fd03276](https://github.com/sumup/sumup-go/commit/fd032762888a0689722c1391f92f309d0185bcbf))
* generate latest sdk ([#16](https://github.com/sumup/sumup-go/issues/16)) ([ed69f6e](https://github.com/sumup/sumup-go/commit/ed69f6ef2e00afab86de8dd4ec6fd3ed16b889ab))
* init ([f0e2412](https://github.com/sumup/sumup-go/commit/f0e2412f7876db07790b531a29f517f092cb33a1))
* releases and changelog ([#27](https://github.com/sumup/sumup-go/issues/27)) ([28f5183](https://github.com/sumup/sumup-go/commit/28f5183ad026a0241d75544a949e5d3ef7ad0500))
* switch to go-sdk-gen ([#20](https://github.com/sumup/sumup-go/issues/20)) ([1223a9f](https://github.com/sumup/sumup-go/commit/1223a9fdd60a3d73f7a064a832834809b0ea0227))
* **tooling:** create releases in draft mode ([#34](https://github.com/sumup/sumup-go/issues/34)) ([6f7c81b](https://github.com/sumup/sumup-go/commit/6f7c81bd0761cd5cb688cd202bef5c38150919f4))
* update to latest specs ([ebdf8a3](https://github.com/sumup/sumup-go/commit/ebdf8a3ea0d9bb4bd8d471873b1ea1a39bc7e84f))
* update to latest version ([#10](https://github.com/sumup/sumup-go/issues/10)) ([b04189a](https://github.com/sumup/sumup-go/commit/b04189a251de09e12cb45213f3eb1dbcea81644c))
* v0.0.1 ([#73](https://github.com/sumup/sumup-go/issues/73)) ([b7a2e03](https://github.com/sumup/sumup-go/commit/b7a2e0313e333450008eea6fa3caef587a13992c))


### Bug Fixes

* **ci:** code generation ([#7](https://github.com/sumup/sumup-go/issues/7)) ([9713af5](https://github.com/sumup/sumup-go/commit/9713af584585652d809177ef01711bad3e8d55dc))
* **ci:** go version for vulncheck ([#28](https://github.com/sumup/sumup-go/issues/28)) ([86c6082](https://github.com/sumup/sumup-go/commit/86c6082a8eaf2570834f139c33f1c82530229896))
* **ci:** release process token permissions ([#30](https://github.com/sumup/sumup-go/issues/30)) ([7442793](https://github.com/sumup/sumup-go/commit/744279379346c04d073fa3e61feeb8be46144e87))
* **ci:** set generate workflow permissions ([#78](https://github.com/sumup/sumup-go/issues/78)) ([e83722f](https://github.com/sumup/sumup-go/commit/e83722ff3774d62e78af8129e0a2985b579a9403))
* **docs:** remove github discussions link from readme ([#25](https://github.com/sumup/sumup-go/issues/25)) ([4b05842](https://github.com/sumup/sumup-go/commit/4b05842b0167c9809b89610359b4030d52756df8))
* **docs:** update readme ([#41](https://github.com/sumup/sumup-go/issues/41)) ([770b300](https://github.com/sumup/sumup-go/commit/770b300ca4841c6fe364f0c44b49f9b3d78457da))
* ReaderService impl ([#9](https://github.com/sumup/sumup-go/issues/9)) ([1a04e19](https://github.com/sumup/sumup-go/commit/1a04e1972645f52dd56e5fe6b9742881e7f83604))
* run latest make generate ([#35](https://github.com/sumup/sumup-go/issues/35)) ([58653cf](https://github.com/sumup/sumup-go/commit/58653cf2e531f12ea8452e4dd8ffc215169f8af1))
* **tooling:** release please config ([#52](https://github.com/sumup/sumup-go/issues/52)) ([ae2aff5](https://github.com/sumup/sumup-go/commit/ae2aff5a95629184bdf8681e8a316e06c207fcdb))
* **tooling:** release please config ([#84](https://github.com/sumup/sumup-go/issues/84)) ([1559ba1](https://github.com/sumup/sumup-go/commit/1559ba17e67b5cadc05d375c8d5064de3342eeb7))
* use idiomatic comment syntax ([186afa2](https://github.com/sumup/sumup-go/commit/186afa25127a1b7154c9f9d51c7bcd7d42746823))

## [0.0.1-beta.4](https://github.com/sumup/sumup-go/compare/v0.0.1-beta.3...v0.0.1-beta.4) (2025-03-04)


### Features

* **cd:** send notification on release ([#32](https://github.com/sumup/sumup-go/issues/32)) ([013a4bb](https://github.com/sumup/sumup-go/commit/013a4bb967730579921e678044fa96fcc8ad48cf))
* **ci/cd:** update actions, add dependabot ([#3](https://github.com/sumup/sumup-go/issues/3)) ([43b02ff](https://github.com/sumup/sumup-go/commit/43b02ff3ba7641e81c2e3bbbf524a8b33b456525))
* **ci:** auto-generate latest SDK ([#5](https://github.com/sumup/sumup-go/issues/5)) ([2343544](https://github.com/sumup/sumup-go/commit/2343544a078da2e6511580048239191ecf008505))
* **ci:** commit generated code ([#6](https://github.com/sumup/sumup-go/issues/6)) ([26cc46b](https://github.com/sumup/sumup-go/commit/26cc46bd644cfd37d81b456ac9babfb680a52d22))
* **ci:** lint github actions ([#22](https://github.com/sumup/sumup-go/issues/22)) ([da1168f](https://github.com/sumup/sumup-go/commit/da1168f1f592af898e3de5ffa992d2e7999b5221))
* doc.go ([44434a7](https://github.com/sumup/sumup-go/commit/44434a7cbf963022c5ddfe7eb85c66069eefb50e))
* **docs:** add link to API reference and developer portal ([#29](https://github.com/sumup/sumup-go/issues/29)) ([ad2507e](https://github.com/sumup/sumup-go/commit/ad2507e4036f058a95b344d268c2a63d9a9f1ecd))
* **docs:** README badges ([28627fb](https://github.com/sumup/sumup-go/commit/28627fb8bebc3bfce09ec39b6b331331112aa18a))
* **docs:** security policy ([fd03276](https://github.com/sumup/sumup-go/commit/fd032762888a0689722c1391f92f309d0185bcbf))
* generate latest sdk ([#16](https://github.com/sumup/sumup-go/issues/16)) ([ed69f6e](https://github.com/sumup/sumup-go/commit/ed69f6ef2e00afab86de8dd4ec6fd3ed16b889ab))
* init ([f0e2412](https://github.com/sumup/sumup-go/commit/f0e2412f7876db07790b531a29f517f092cb33a1))
* releases and changelog ([#27](https://github.com/sumup/sumup-go/issues/27)) ([28f5183](https://github.com/sumup/sumup-go/commit/28f5183ad026a0241d75544a949e5d3ef7ad0500))
* switch to go-sdk-gen ([#20](https://github.com/sumup/sumup-go/issues/20)) ([1223a9f](https://github.com/sumup/sumup-go/commit/1223a9fdd60a3d73f7a064a832834809b0ea0227))
* **tooling:** create releases in draft mode ([#34](https://github.com/sumup/sumup-go/issues/34)) ([6f7c81b](https://github.com/sumup/sumup-go/commit/6f7c81bd0761cd5cb688cd202bef5c38150919f4))
* update to latest specs ([ebdf8a3](https://github.com/sumup/sumup-go/commit/ebdf8a3ea0d9bb4bd8d471873b1ea1a39bc7e84f))
* update to latest version ([#10](https://github.com/sumup/sumup-go/issues/10)) ([b04189a](https://github.com/sumup/sumup-go/commit/b04189a251de09e12cb45213f3eb1dbcea81644c))


### Bug Fixes

* **ci:** code generation ([#7](https://github.com/sumup/sumup-go/issues/7)) ([9713af5](https://github.com/sumup/sumup-go/commit/9713af584585652d809177ef01711bad3e8d55dc))
* **ci:** go version for vulncheck ([#28](https://github.com/sumup/sumup-go/issues/28)) ([86c6082](https://github.com/sumup/sumup-go/commit/86c6082a8eaf2570834f139c33f1c82530229896))
* **ci:** release process token permissions ([#30](https://github.com/sumup/sumup-go/issues/30)) ([7442793](https://github.com/sumup/sumup-go/commit/744279379346c04d073fa3e61feeb8be46144e87))
* **docs:** remove github discussions link from readme ([#25](https://github.com/sumup/sumup-go/issues/25)) ([4b05842](https://github.com/sumup/sumup-go/commit/4b05842b0167c9809b89610359b4030d52756df8))
* **docs:** update readme ([#41](https://github.com/sumup/sumup-go/issues/41)) ([770b300](https://github.com/sumup/sumup-go/commit/770b300ca4841c6fe364f0c44b49f9b3d78457da))
* ReaderService impl ([#9](https://github.com/sumup/sumup-go/issues/9)) ([1a04e19](https://github.com/sumup/sumup-go/commit/1a04e1972645f52dd56e5fe6b9742881e7f83604))
* run latest make generate ([#35](https://github.com/sumup/sumup-go/issues/35)) ([58653cf](https://github.com/sumup/sumup-go/commit/58653cf2e531f12ea8452e4dd8ffc215169f8af1))
* **tooling:** release please config ([#52](https://github.com/sumup/sumup-go/issues/52)) ([ae2aff5](https://github.com/sumup/sumup-go/commit/ae2aff5a95629184bdf8681e8a316e06c207fcdb))
* use idiomatic comment syntax ([186afa2](https://github.com/sumup/sumup-go/commit/186afa25127a1b7154c9f9d51c7bcd7d42746823))

## [0.0.1-beta.3](https://github.com/sumup/sumup-go/compare/v0.0.1-beta.2...v0.0.1-beta.3) (2025-03-04)


### Bug Fixes

* **tooling:** release please config ([#52](https://github.com/sumup/sumup-go/issues/52)) ([ae2aff5](https://github.com/sumup/sumup-go/commit/ae2aff5a95629184bdf8681e8a316e06c207fcdb))

## [0.0.1-beta.2](https://github.com/sumup/sumup-go/compare/v0.0.1-beta.1...v0.0.1-beta.2) (2025-02-16)


### Features

* **cd:** send notification on release ([#32](https://github.com/sumup/sumup-go/issues/32)) ([013a4bb](https://github.com/sumup/sumup-go/commit/013a4bb967730579921e678044fa96fcc8ad48cf))
* **ci/cd:** update actions, add dependabot ([#3](https://github.com/sumup/sumup-go/issues/3)) ([43b02ff](https://github.com/sumup/sumup-go/commit/43b02ff3ba7641e81c2e3bbbf524a8b33b456525))
* **ci:** auto-generate latest SDK ([#5](https://github.com/sumup/sumup-go/issues/5)) ([2343544](https://github.com/sumup/sumup-go/commit/2343544a078da2e6511580048239191ecf008505))
* **ci:** commit generated code ([#6](https://github.com/sumup/sumup-go/issues/6)) ([26cc46b](https://github.com/sumup/sumup-go/commit/26cc46bd644cfd37d81b456ac9babfb680a52d22))
* **ci:** lint github actions ([#22](https://github.com/sumup/sumup-go/issues/22)) ([da1168f](https://github.com/sumup/sumup-go/commit/da1168f1f592af898e3de5ffa992d2e7999b5221))
* doc.go ([44434a7](https://github.com/sumup/sumup-go/commit/44434a7cbf963022c5ddfe7eb85c66069eefb50e))
* **docs:** add link to API reference and developer portal ([#29](https://github.com/sumup/sumup-go/issues/29)) ([ad2507e](https://github.com/sumup/sumup-go/commit/ad2507e4036f058a95b344d268c2a63d9a9f1ecd))
* **docs:** README badges ([28627fb](https://github.com/sumup/sumup-go/commit/28627fb8bebc3bfce09ec39b6b331331112aa18a))
* **docs:** security policy ([fd03276](https://github.com/sumup/sumup-go/commit/fd032762888a0689722c1391f92f309d0185bcbf))
* generate latest sdk ([#16](https://github.com/sumup/sumup-go/issues/16)) ([ed69f6e](https://github.com/sumup/sumup-go/commit/ed69f6ef2e00afab86de8dd4ec6fd3ed16b889ab))
* init ([f0e2412](https://github.com/sumup/sumup-go/commit/f0e2412f7876db07790b531a29f517f092cb33a1))
* releases and changelog ([#27](https://github.com/sumup/sumup-go/issues/27)) ([28f5183](https://github.com/sumup/sumup-go/commit/28f5183ad026a0241d75544a949e5d3ef7ad0500))
* switch to go-sdk-gen ([#20](https://github.com/sumup/sumup-go/issues/20)) ([1223a9f](https://github.com/sumup/sumup-go/commit/1223a9fdd60a3d73f7a064a832834809b0ea0227))
* **tooling:** create releases in draft mode ([#34](https://github.com/sumup/sumup-go/issues/34)) ([6f7c81b](https://github.com/sumup/sumup-go/commit/6f7c81bd0761cd5cb688cd202bef5c38150919f4))
* update to latest specs ([ebdf8a3](https://github.com/sumup/sumup-go/commit/ebdf8a3ea0d9bb4bd8d471873b1ea1a39bc7e84f))
* update to latest version ([#10](https://github.com/sumup/sumup-go/issues/10)) ([b04189a](https://github.com/sumup/sumup-go/commit/b04189a251de09e12cb45213f3eb1dbcea81644c))


### Bug Fixes

* **ci:** code generation ([#7](https://github.com/sumup/sumup-go/issues/7)) ([9713af5](https://github.com/sumup/sumup-go/commit/9713af584585652d809177ef01711bad3e8d55dc))
* **ci:** go version for vulncheck ([#28](https://github.com/sumup/sumup-go/issues/28)) ([86c6082](https://github.com/sumup/sumup-go/commit/86c6082a8eaf2570834f139c33f1c82530229896))
* **ci:** release process token permissions ([#30](https://github.com/sumup/sumup-go/issues/30)) ([7442793](https://github.com/sumup/sumup-go/commit/744279379346c04d073fa3e61feeb8be46144e87))
* **docs:** remove github discussions link from readme ([#25](https://github.com/sumup/sumup-go/issues/25)) ([4b05842](https://github.com/sumup/sumup-go/commit/4b05842b0167c9809b89610359b4030d52756df8))
* **docs:** update readme ([#41](https://github.com/sumup/sumup-go/issues/41)) ([770b300](https://github.com/sumup/sumup-go/commit/770b300ca4841c6fe364f0c44b49f9b3d78457da))
* ReaderService impl ([#9](https://github.com/sumup/sumup-go/issues/9)) ([1a04e19](https://github.com/sumup/sumup-go/commit/1a04e1972645f52dd56e5fe6b9742881e7f83604))
* run latest make generate ([#35](https://github.com/sumup/sumup-go/issues/35)) ([58653cf](https://github.com/sumup/sumup-go/commit/58653cf2e531f12ea8452e4dd8ffc215169f8af1))
* use idiomatic comment syntax ([186afa2](https://github.com/sumup/sumup-go/commit/186afa25127a1b7154c9f9d51c7bcd7d42746823))

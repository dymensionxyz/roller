# Changelog

## [1.7.0-alpha-pre-rc](https://github.com/dymensionxyz/roller/compare/v1.7.1-alpha-rc01...v1.7.0-alpha-pre-rc) (2024-09-19)


### ⚠ BREAKING CHANGES

* Add support for standard ibc ([#614](https://github.com/dymensionxyz/roller/issues/614))

### Features

* `roller config init` command output formatting ([#23](https://github.com/dymensionxyz/roller/issues/23)) ([9b638a6](https://github.com/dymensionxyz/roller/commit/9b638a6f24cb9c118f1582a344b89197932a0749))
* `roller config init` light node configuration ([#13](https://github.com/dymensionxyz/roller/issues/13)) ([1a7f4cc](https://github.com/dymensionxyz/roller/commit/1a7f4cc2d2bee5a6f257ae34a72504c00e96a80b))
* accepts -i for interactive. arguments are mandatory ([#186](https://github.com/dymensionxyz/roller/issues/186)) ([16317b6](https://github.com/dymensionxyz/roller/commit/16317b618665935c2dbe23793aac62d2672bcdab))
* Add 'roller services load' command to load the different RollApp services ([#282](https://github.com/dymensionxyz/roller/issues/282)) ([15bd450](https://github.com/dymensionxyz/roller/commit/15bd4507a9f7ae16488371f851baa0f14ca6b9bb))
* add `--mock` flag to init ([#912](https://github.com/dymensionxyz/roller/issues/912)) ([0fccdb3](https://github.com/dymensionxyz/roller/commit/0fccdb3cb4b15b69ead383bad6c20493ab1f4f82))
* Add `roller config set hub-rpc` ([#532](https://github.com/dymensionxyz/roller/issues/532)) ([1ace736](https://github.com/dymensionxyz/roller/commit/1ace736d304dea3e5bf0b87f6a6372d034ff6743))
* Add a `claim-rewards` command ([#566](https://github.com/dymensionxyz/roller/issues/566)) ([dffb1ec](https://github.com/dymensionxyz/roller/commit/dffb1ec5db2f1aa645e9984685fb4aec0610edb6))
* Add a `no-output` flag to `roller run` ([#400](https://github.com/dymensionxyz/roller/issues/400)) ([4aea642](https://github.com/dymensionxyz/roller/commit/4aea642d2d221d2ade6583a5a6059339eb8f3385))
* Add a `no-output` flag to roller init and register ([#406](https://github.com/dymensionxyz/roller/issues/406)) ([7624ca9](https://github.com/dymensionxyz/roller/commit/7624ca98fbdcf5316b405f8e9bc5dd5c3a14cdd9))
* Add a `roller config export` command ([#439](https://github.com/dymensionxyz/roller/issues/439)) ([c38631b](https://github.com/dymensionxyz/roller/commit/c38631b6322e2c85d9ebdb79950ef29190c48d80))
* Add a `roller config show` command ([#351](https://github.com/dymensionxyz/roller/issues/351)) ([6faff32](https://github.com/dymensionxyz/roller/commit/6faff321c1cb1f25e12ba939532211a0748c0e5a))
* Add a `roller relayer channel show` command ([#352](https://github.com/dymensionxyz/roller/issues/352)) ([1d8d7e6](https://github.com/dymensionxyz/roller/commit/1d8d7e6aab1832f8fc069b46744495baad847389))
* Add a `roller tx fund-faucet` command ([#479](https://github.com/dymensionxyz/roller/issues/479)) ([9eb65b0](https://github.com/dymensionxyz/roller/commit/9eb65b08a925ce1e6ee2d82aa85657f65174d7a4))
* Add a command to change DA with an existing rollapp ([#451](https://github.com/dymensionxyz/roller/issues/451)) ([487f0c8](https://github.com/dymensionxyz/roller/commit/487f0c8d2f55375db4bc915206d7b3f6155515ea))
* Add a configurable decimals flag to `roller config init` ([#152](https://github.com/dymensionxyz/roller/issues/152)) ([f77f605](https://github.com/dymensionxyz/roller/commit/f77f605406625ff8ff25aa40be5075bf8040ceb2))
* Add a sequencer status command ([#431](https://github.com/dymensionxyz/roller/issues/431)) ([764f1eb](https://github.com/dymensionxyz/roller/commit/764f1eb4dc861180415f7bb027901bcf860594e3))
* Add ability to list roller addresses with `roller keys list` command ([#83](https://github.com/dymensionxyz/roller/issues/83)) ([b4c875e](https://github.com/dymensionxyz/roller/commit/b4c875e2508713204adf6c3507713903bc1258d9))
* Add ability to run Celestia DA light node ([#71](https://github.com/dymensionxyz/roller/issues/71)) ([eb18261](https://github.com/dymensionxyz/roller/commit/eb18261c620f4af8f78a54f775f0b4ea59d26295))
* Add ability to run rollapps with the `roller run` command ([#66](https://github.com/dymensionxyz/roller/issues/66)) ([9675558](https://github.com/dymensionxyz/roller/commit/96755584d993bcacdf49e40b59cf00a7ad83f74b))
* Add ability to run the relayer ([#78](https://github.com/dymensionxyz/roller/issues/78)) ([cada290](https://github.com/dymensionxyz/roller/commit/cada29074c0159963fb7e36d6eb19b3f57548538))
* Add ability to switch between different hub environments ([#55](https://github.com/dymensionxyz/roller/issues/55)) ([862aecb](https://github.com/dymensionxyz/roller/commit/862aecb7e06778e1c77a4f61db22cba308f281d7))
* Add automatic rollapp relayer funding in the genesis file ([#95](https://github.com/dymensionxyz/roller/issues/95)) ([997af61](https://github.com/dymensionxyz/roller/commit/997af6156c606064caa35e9126704a3cfdd74f36))
* Add Balance Verification for Balance-Required Commands ([#86](https://github.com/dymensionxyz/roller/issues/86)) ([fe5b18f](https://github.com/dymensionxyz/roller/commit/fe5b18fa33d83af7054073a0f06510691fb3ec26))
* add bank denom metadata to rollapp genesis ([#292](https://github.com/dymensionxyz/roller/issues/292)) ([885ea2b](https://github.com/dymensionxyz/roller/commit/885ea2b22f86476d5d4c6691f315eb8dd3d8ac6c))
* add binary installation command ([#879](https://github.com/dymensionxyz/roller/issues/879)) ([71657bd](https://github.com/dymensionxyz/roller/commit/71657bd8a5cb23196968aa683b0e3c88e56fc5b4))
* add block explorer command ([#880](https://github.com/dymensionxyz/roller/issues/880)) ([73e0cb2](https://github.com/dymensionxyz/roller/commit/73e0cb2845014b8f812efc238a3d40ef8dc5406b))
* Add celestia balance verification ([#89](https://github.com/dymensionxyz/roller/issues/89)) ([7a82e7e](https://github.com/dymensionxyz/roller/commit/7a82e7e882f060520a0d7c450bb00161e29ac4ba))
* Add clear output to the roller run command ([#68](https://github.com/dymensionxyz/roller/issues/68)) ([825a385](https://github.com/dymensionxyz/roller/commit/825a385085fbec3fe9e234480a5cc3ac47e592fe))
* Add command to modify DA light client rpc port ([#369](https://github.com/dymensionxyz/roller/issues/369)) ([c082e0d](https://github.com/dymensionxyz/roller/commit/c082e0da5af13e4c331c19d5a1eca2ad8523137a))
* Add command to modify rollapp rpc port ([#365](https://github.com/dymensionxyz/roller/issues/365)) ([2c2ad89](https://github.com/dymensionxyz/roller/commit/2c2ad89ff132764daf1204031ac471a9bf491a79))
* Add command to register rollapps ([#50](https://github.com/dymensionxyz/roller/issues/50)) ([aee609f](https://github.com/dymensionxyz/roller/commit/aee609f26b8fcb0d96dfef8a09923465e71d0a50))
* add commands to export and update sequencer metadata ([#847](https://github.com/dymensionxyz/roller/issues/847)) ([3febe76](https://github.com/dymensionxyz/roller/commit/3febe76fea22c2afa994c113a790d57a92597db4))
* Add comprehensive run command ([#118](https://github.com/dymensionxyz/roller/issues/118)) ([a9f8165](https://github.com/dymensionxyz/roller/commit/a9f8165e834c87c77c46440c813856c92e41eb56))
* Add configurable `token-supply` flag on `roller config init` ([#94](https://github.com/dymensionxyz/roller/issues/94)) ([e8344e4](https://github.com/dymensionxyz/roller/commit/e8344e48f6e88b09844fde7f042775c50d28d90b))
* Add Continuous Status Monitoring for DA LC and sequencer ([#137](https://github.com/dymensionxyz/roller/issues/137)) ([de39ccb](https://github.com/dymensionxyz/roller/commit/de39ccb43ef6d66cf526b6aaf8887fab1a8f825c))
* add cron job for relayer flush ([#954](https://github.com/dymensionxyz/roller/issues/954)) ([24d39dd](https://github.com/dymensionxyz/roller/commit/24d39ddecdff423ec6084899f83eb12905a437c2))
* add eibc client support ([#813](https://github.com/dymensionxyz/roller/issues/813)) ([89dc1c6](https://github.com/dymensionxyz/roller/commit/89dc1c67bcb54510f44bf17eeb4cfd634dd25a7d))
* Add IBC channel creation capability from `roller relayer start` command ([#74](https://github.com/dymensionxyz/roller/issues/74)) ([8619333](https://github.com/dymensionxyz/roller/commit/86193333d1fb78d1af22841b3ecbc0e1af36d0c4))
* Add installation script for latest release fetching ([#271](https://github.com/dymensionxyz/roller/issues/271)) ([fc033e0](https://github.com/dymensionxyz/roller/commit/fc033e081beb6cb53ab60c5520b1060baaa06693))
* add interactive cli for rollapp init ([#136](https://github.com/dymensionxyz/roller/issues/136)) ([a6f207a](https://github.com/dymensionxyz/roller/commit/a6f207ad1b0d6754d783bdae0b3348d0356f7b78))
* Add keys export command ([#143](https://github.com/dymensionxyz/roller/issues/143)) ([998f4e4](https://github.com/dymensionxyz/roller/commit/998f4e493930b456a1e4b49abab9a7aac9dcbb05))
* add launchctl service support for macos ([#951](https://github.com/dymensionxyz/roller/issues/951)) ([f518849](https://github.com/dymensionxyz/roller/commit/f5188497c9b6abd489a3d72289b4e2fd2c0e463a))
* Add loading spinner with status on roller init, register and run ([#166](https://github.com/dymensionxyz/roller/issues/166)) ([63d166d](https://github.com/dymensionxyz/roller/commit/63d166d1021c64a41837b69898086108a5ea8f3f))
* Add local hub initialization logic to roller config init ([#381](https://github.com/dymensionxyz/roller/issues/381)) ([9f896f7](https://github.com/dymensionxyz/roller/commit/9f896f7d0c1a7046dc7f50ee700cc57157e857ee))
* add logs command for services ([#943](https://github.com/dymensionxyz/roller/issues/943)) ([360644f](https://github.com/dymensionxyz/roller/commit/360644f64bc65ae65c44d64c8a63cf95f52fbefa))
* Add option to export Celestia private key ([#242](https://github.com/dymensionxyz/roller/issues/242)) ([bda8916](https://github.com/dymensionxyz/roller/commit/bda891649c3e361651c00ea21a4c4aed06872988))
* Add output formatting for `roller register` command ([#62](https://github.com/dymensionxyz/roller/issues/62)) ([eaa2728](https://github.com/dymensionxyz/roller/commit/eaa27286b8c9c8fac9b5bfe0b91819bf238c0427))
* add progress bar to downloads ([#965](https://github.com/dymensionxyz/roller/issues/965)) ([e6d3b4b](https://github.com/dymensionxyz/roller/commit/e6d3b4b8ce11783e55a96b9e85a0c41414cc2954))
* Add rollapp height to run output ([#218](https://github.com/dymensionxyz/roller/issues/218)) ([a6eca00](https://github.com/dymensionxyz/roller/commit/a6eca007578311f0ac89af6ad6b2ce6897ccae01))
* add rollapp init ([#803](https://github.com/dymensionxyz/roller/issues/803)) ([19a455a](https://github.com/dymensionxyz/roller/commit/19a455a0d30c58b424408b80741ba7f2a21cf027))
* add rollapp status command ([#812](https://github.com/dymensionxyz/roller/issues/812)) ([0e1d2a1](https://github.com/dymensionxyz/roller/commit/0e1d2a1ee7cb17f5586e81693ebdc0643106a475))
* Add roller migrate for Configuration Upgrades ([#317](https://github.com/dymensionxyz/roller/issues/317)) ([7890257](https://github.com/dymensionxyz/roller/commit/78902574620b46dbe931e72dc0701542de9bbcbc))
* add sequencer address and balance to `rollapp` output commands ([#948](https://github.com/dymensionxyz/roller/issues/948)) ([7668320](https://github.com/dymensionxyz/roller/commit/7668320ca951942403baa9dda5204e3751a1738f))
* Add sequencer registration on `roller register` ([#65](https://github.com/dymensionxyz/roller/issues/65)) ([870b254](https://github.com/dymensionxyz/roller/commit/870b2546148357c0546c49b1eabdee60246b5b26))
* Add services log files and ports to roller run output ([#145](https://github.com/dymensionxyz/roller/issues/145)) ([82b4d41](https://github.com/dymensionxyz/roller/commit/82b4d4161db0c268e14aa7ab172836c4c9dc8c3c))
* Add set functions to all rollapp ports ([#457](https://github.com/dymensionxyz/roller/issues/457)) ([8171db4](https://github.com/dymensionxyz/roller/commit/8171db4efcb6e80bfd7fa8f0b23e38deff8e7e14))
* Add support for a custom RDK binary ([#181](https://github.com/dymensionxyz/roller/issues/181)) ([9c9fd55](https://github.com/dymensionxyz/roller/commit/9c9fd550ce263ed13eaeea636bf7bbb37e5cb9af))
* add support for avail as a da ([#285](https://github.com/dymensionxyz/roller/issues/285)) ([d2b3500](https://github.com/dymensionxyz/roller/commit/d2b35004fdf97a49990792b48eeeebb49f8adc85))
* add support for mock da ([#180](https://github.com/dymensionxyz/roller/issues/180)) ([8b50d34](https://github.com/dymensionxyz/roller/commit/8b50d34c8dd3dba7e59ccf1be9242a269ffcfd7e))
* Add support for setting sequencer as sole governer by default ([#376](https://github.com/dymensionxyz/roller/issues/376)) ([b2ea7e6](https://github.com/dymensionxyz/roller/commit/b2ea7e686e57570797428e848eba9b9b2a57f571))
* Add support for standard ibc ([#614](https://github.com/dymensionxyz/roller/issues/614)) ([3be9ca6](https://github.com/dymensionxyz/roller/commit/3be9ca6177456a645d7fd0966765e69358605806))
* add wasm rollapp support ([#968](https://github.com/dymensionxyz/roller/issues/968)) ([dc57755](https://github.com/dymensionxyz/roller/commit/dc57755a8574ec817cf40eba35d7647720974c79))
* Added configuration files generation ability to roller init ([#8](https://github.com/dymensionxyz/roller/issues/8)) ([de8f0a7](https://github.com/dymensionxyz/roller/commit/de8f0a7c3c03af7be9b03385a02c6977ef67127b))
* added gentx_seq call on init ([#156](https://github.com/dymensionxyz/roller/issues/156)) ([e327098](https://github.com/dymensionxyz/roller/commit/e327098b5fc0d02cb3f5512112fca8d9a7dcf2dd))
* Added installation script for mac os ([#35](https://github.com/dymensionxyz/roller/issues/35)) ([77efd90](https://github.com/dymensionxyz/roller/commit/77efd9006040b73642bff7b1a48247c2cf2cffb0))
* Added key generation ability to roller init ([#6](https://github.com/dymensionxyz/roller/issues/6)) ([de1e2c0](https://github.com/dymensionxyz/roller/commit/de1e2c0ae43f4d191e063e18cdcef5c043144c19))
* added Makefile with build and install targets ([#131](https://github.com/dymensionxyz/roller/issues/131)) ([d3f6ead](https://github.com/dymensionxyz/roller/commit/d3f6ead7b0c80e4a3bb5e077db1c678353f32fa4))
* Added root command ([#2](https://github.com/dymensionxyz/roller/issues/2)) ([e3a93ed](https://github.com/dymensionxyz/roller/commit/e3a93ed0e43ffd95b637cb2a4764b870a63f67da))
* Added version command ([#37](https://github.com/dymensionxyz/roller/issues/37)) ([c406eb7](https://github.com/dymensionxyz/roller/commit/c406eb7a7e9651d509cb1d560c5a1faea37ba255))
* added vm type parameter ([#319](https://github.com/dymensionxyz/roller/issues/319)) ([a40a499](https://github.com/dymensionxyz/roller/commit/a40a499eaf21ee5f5044f55eba8613d609f853ac))
* adding --force flag for sequencer registration ([#203](https://github.com/dymensionxyz/roller/issues/203)) ([b6fbb74](https://github.com/dymensionxyz/roller/commit/b6fbb74c7b606e8db286b8eb96c9fb67e1be7646))
* Auto-generate EIP115 and revision ([#499](https://github.com/dymensionxyz/roller/issues/499)) ([7715895](https://github.com/dymensionxyz/roller/commit/7715895b6cf5f65caddb44e1eacf8de47467fa97))
* Automatically Create and Upload Assets for New Releases ([#268](https://github.com/dymensionxyz/roller/issues/268)) ([0318e54](https://github.com/dymensionxyz/roller/commit/0318e54d43de54f91dc3d5aa1d20241aefe2567f))
* Avail - Support for Exporting Private Key ([#349](https://github.com/dymensionxyz/roller/issues/349)) ([21c4ff4](https://github.com/dymensionxyz/roller/commit/21c4ff4517b5c144ba2f5c047fd9815daf97c7a9))
* **block-explorer:** add a command to tear down the block explorer setup ([#979](https://github.com/dymensionxyz/roller/issues/979)) ([4890bc7](https://github.com/dymensionxyz/roller/commit/4890bc7b1fb09070bb9abce362affb86fba3e7fd))
* celestia light client config logic, sync from da head for first sequencer, sync from first state update for others ([#832](https://github.com/dymensionxyz/roller/issues/832)) ([c911b45](https://github.com/dymensionxyz/roller/commit/c911b451d90c468a739692e2d182fd2023694669))
* Configurable root directory in `roller config init` ([#25](https://github.com/dymensionxyz/roller/issues/25)) ([c89328b](https://github.com/dymensionxyz/roller/commit/c89328b2e0afb3088df5a1e3ed271302a5191ba0))
* denom accepts 3-6 letters ([#200](https://github.com/dymensionxyz/roller/issues/200)) ([3417ceb](https://github.com/dymensionxyz/roller/commit/3417ceb8546ac5c9e061a6e63602f08cd1498673))
* display and accept bond amount in base denom ([#911](https://github.com/dymensionxyz/roller/issues/911)) ([61b445b](https://github.com/dymensionxyz/roller/commit/61b445b210f51ebb08ad9917a2805e7528f30648))
* Enable tag-specific roller installation  ([#411](https://github.com/dymensionxyz/roller/issues/411)) ([ee6cccf](https://github.com/dymensionxyz/roller/commit/ee6cccf8337da5bd26d4b631eafc23b448ae30c0))
* fetch DA information from `roller.toml` for easy endpoint changes ([#862](https://github.com/dymensionxyz/roller/issues/862)) ([1d66c64](https://github.com/dymensionxyz/roller/commit/1d66c64d1b7e3c895fb7f465eadee5fb4f4f9253))
* fetch rollapp info from chain ([#828](https://github.com/dymensionxyz/roller/issues/828)) ([130eb89](https://github.com/dymensionxyz/roller/commit/130eb8997da75559afd1255048b00f86ba7a9473))
* finalize full node flow ([#836](https://github.com/dymensionxyz/roller/issues/836)) ([ae3000b](https://github.com/dymensionxyz/roller/commit/ae3000bb6d42854181274e5c7263627eee48931b))
* finalize relayer flow ([#833](https://github.com/dymensionxyz/roller/issues/833)) ([ef430b1](https://github.com/dymensionxyz/roller/commit/ef430b1b460c8858d65271498bf36581355d3583))
* finalize sequencer registration flow ([#831](https://github.com/dymensionxyz/roller/issues/831)) ([3c0296a](https://github.com/dymensionxyz/roller/commit/3c0296a3b5c807a6fcb7665c9f32f0e3c9369eaa))
* fix balance printing in roller run ([#193](https://github.com/dymensionxyz/roller/issues/193)) ([969670b](https://github.com/dymensionxyz/roller/commit/969670b23aeca4692bdae3a04cd2ab55568c86de))
* full node flow ([#827](https://github.com/dymensionxyz/roller/issues/827)) ([95a9bff](https://github.com/dymensionxyz/roller/commit/95a9bff4263efdf57f5a471022ceee81598e8a98))
* generate chain config during block-explorer initialization ([#884](https://github.com/dymensionxyz/roller/issues/884)) ([c8afc46](https://github.com/dymensionxyz/roller/commit/c8afc46267670485a579be336a99b4e7f88520de))
* handle genesis creator archive ([#800](https://github.com/dymensionxyz/roller/issues/800)) ([e3ea97c](https://github.com/dymensionxyz/roller/commit/e3ea97c123a68ffa7937199fc81876b25dfc6494))
* handle sequencer bond ([#906](https://github.com/dymensionxyz/roller/issues/906)) ([c18e989](https://github.com/dymensionxyz/roller/commit/c18e9890b8a7f9c58299b420e0bc3f3915df24ac))
* handle sequencer reward address ([#866](https://github.com/dymensionxyz/roller/issues/866)) ([87f3765](https://github.com/dymensionxyz/roller/commit/87f37659c1e64e6697039358a7c602d031230615))
* Implemented rerun handling logic for roller config init ([#31](https://github.com/dymensionxyz/roller/issues/31)) ([8171c61](https://github.com/dymensionxyz/roller/commit/8171c61bc0eebbf0b81f18bb1b9bbcb1b747f9dd))
* improve rollapp initialization ([#817](https://github.com/dymensionxyz/roller/issues/817)) ([d7db004](https://github.com/dymensionxyz/roller/commit/d7db0042ba18d342f92ed160fd8f18659dbb7186))
* Initializing relayer configuration from `roller config init` ([#19](https://github.com/dymensionxyz/roller/issues/19)) ([c934949](https://github.com/dymensionxyz/roller/commit/c93494968deda9c6d332df0276b298380f81e17d))
* install binaries based on rollapp id ([#885](https://github.com/dymensionxyz/roller/issues/885)) ([ce93e8d](https://github.com/dymensionxyz/roller/commit/ce93e8dca399335157d808327bb2764e4cec2563))
* install script compiles all dependencies locally ([#211](https://github.com/dymensionxyz/roller/issues/211)) ([c1c4803](https://github.com/dymensionxyz/roller/commit/c1c4803a9f726252ed8152091d0f6d57ffdf0828))
* make it possible to fill out the moniker and social media links for sequencer ([#923](https://github.com/dymensionxyz/roller/issues/923)) ([a8cf922](https://github.com/dymensionxyz/roller/commit/a8cf9221e5b08c85700b58368f46f16d7dbdc967))
* Output of DA Light Client and Relayer written to dedicated log files ([#91](https://github.com/dymensionxyz/roller/issues/91)) ([ae57f48](https://github.com/dymensionxyz/roller/commit/ae57f48965e19e4a27b6f367662e662178dcc7db))
* removed prompt from installation script ([#42](https://github.com/dymensionxyz/roller/issues/42)) ([72ed72d](https://github.com/dymensionxyz/roller/commit/72ed72dfc78b200bcbd61f3330e84acc794c9180))
* replaced celestia default da network of arabica with mocha ([#161](https://github.com/dymensionxyz/roller/issues/161)) ([a15286f](https://github.com/dymensionxyz/roller/commit/a15286f87908a9773706ec6b2575a25dc49edf6e))
* rollapp id verification ([#48](https://github.com/dymensionxyz/roller/issues/48)) ([3b171b7](https://github.com/dymensionxyz/roller/commit/3b171b7d4e51407bff5c466c5c759d247d60f634))
* roller service output should be generalised for da layer ([#184](https://github.com/dymensionxyz/roller/issues/184)) ([9e27d18](https://github.com/dymensionxyz/roller/commit/9e27d184f6e2054859fc51881aad45a5395ede3b))
* run relayer using roller ([#819](https://github.com/dymensionxyz/roller/issues/819)) ([cf84f05](https://github.com/dymensionxyz/roller/commit/cf84f055e0635abeb349dcb396f6315e6f608a90))
* separate service command into rollapp specific services ([3a8c36f](https://github.com/dymensionxyz/roller/commit/3a8c36f7f74dae64e6a3dbe95abd902eb10aecd6))
* sequencer registration ([#821](https://github.com/dymensionxyz/roller/issues/821)) ([8e7fe60](https://github.com/dymensionxyz/roller/commit/8e7fe60b52d03614c2c206518caec70921814fc6))
* stabilize relayer flow ([#308](https://github.com/dymensionxyz/roller/issues/308)) ([c0648cc](https://github.com/dymensionxyz/roller/commit/c0648ccfd697d4e745f55ae2352eb48169255cd8))
* support latest RDK on devnet ([#132](https://github.com/dymensionxyz/roller/issues/132)) ([17698eb](https://github.com/dymensionxyz/roller/commit/17698eb61878e065c0f88375e820d1fdd96e2488))
* support sequencer bond for registration ([#762](https://github.com/dymensionxyz/roller/issues/762)) ([1ca9f55](https://github.com/dymensionxyz/roller/commit/1ca9f559f673fcdcfce29762696ae4f60722e79d))
* update relayer flow to support lc ([#875](https://github.com/dymensionxyz/roller/issues/875)) ([4d494d5](https://github.com/dymensionxyz/roller/commit/4d494d5915b832b4aa6bb8346c0fe7a48038b65f))
* update sequencer metadata ([#850](https://github.com/dymensionxyz/roller/issues/850)) ([2166bc6](https://github.com/dymensionxyz/roller/commit/2166bc6095736a8343b6bdb161dd66202e03d114))
* using flags as defualts in interactive mode ([#528](https://github.com/dymensionxyz/roller/issues/528)) ([25f0a12](https://github.com/dymensionxyz/roller/commit/25f0a12e15e608aa2ff11f1dc278247a1b00e6d3))
* validate rollapp ID before dependency installation ([#919](https://github.com/dymensionxyz/roller/issues/919)) ([bcf59eb](https://github.com/dymensionxyz/roller/commit/bcf59eb987bad9f280a97e46217e1ceee94e70d7))
* wrap eibc version with min fee percentage support ([#874](https://github.com/dymensionxyz/roller/issues/874)) ([ae9432f](https://github.com/dymensionxyz/roller/commit/ae9432faa332aba64834f8aeae31d5d6f1480124))


### Bug Fixes

* Add a migration to update the relayer path ([#482](https://github.com/dymensionxyz/roller/issues/482)) ([92e36c3](https://github.com/dymensionxyz/roller/commit/92e36c3c5901064e32158fbf248f4cfe69274083))
* add genesis account to genesis file ([#815](https://github.com/dymensionxyz/roller/issues/815)) ([2eb6a09](https://github.com/dymensionxyz/roller/commit/2eb6a090d4438e2bf1a33477866a2dfcf7d0f619))
* Add rollapp ID `xxxx_num_num` enforcement on `roller config init` ([#84](https://github.com/dymensionxyz/roller/issues/84)) ([3eb406f](https://github.com/dymensionxyz/roller/commit/3eb406f08a3467e35645282ec88ca3f2230c5024))
* add sequencer addresses to the generated genesis file ([#808](https://github.com/dymensionxyz/roller/issues/808)) ([01827f4](https://github.com/dymensionxyz/roller/commit/01827f4bcdb566adb8b3710a2b23ccc96b101331))
* Added migration to change hub rpc in relayer to point to archive node ([#578](https://github.com/dymensionxyz/roller/issues/578)) ([f02b038](https://github.com/dymensionxyz/roller/commit/f02b0386425b2fc2813688878c8db3a773230fdd))
* added update clients command after client creation ([#142](https://github.com/dymensionxyz/roller/issues/142)) ([3b99d74](https://github.com/dymensionxyz/roller/commit/3b99d7470d8042df77d37df77a3f473b471d42eb))
* adjusted timeouts. added retry loop over create channels ([#426](https://github.com/dymensionxyz/roller/issues/426)) ([1b44e2d](https://github.com/dymensionxyz/roller/commit/1b44e2d489320a163fbdf0baa343db1da8274bab))
* Avoid panic on failure to fetch da account data ([#232](https://github.com/dymensionxyz/roller/issues/232)) ([b014553](https://github.com/dymensionxyz/roller/commit/b0145537a94a9df9265dda6be94599504c6f5cac))
* **block-explorer:** make the block-explorer creation more reliable ([#977](https://github.com/dymensionxyz/roller/issues/977)) ([fb69d9d](https://github.com/dymensionxyz/roller/commit/fb69d9d3fde392f13ae78ae58e39894f87b43e79))
* Build CD flow roller from the release tag ([#277](https://github.com/dymensionxyz/roller/issues/277)) ([56ac9b7](https://github.com/dymensionxyz/roller/commit/56ac9b7f22ce6f54d7f81bd88dfe69028d58a0bf))
* Celestia LC status being fetched from local LC ([#216](https://github.com/dymensionxyz/roller/issues/216)) ([90ab4ea](https://github.com/dymensionxyz/roller/commit/90ab4ea567fad986848b1be7099de41092c98508))
* celestia namespace retrieval ([a0dd6f6](https://github.com/dymensionxyz/roller/commit/a0dd6f6f6efce445a109e71380dbc272d4b91666))
* change default confirm value to false for all interactions ([#921](https://github.com/dymensionxyz/roller/issues/921)) ([c5d3912](https://github.com/dymensionxyz/roller/commit/c5d391207d82190d21af71ef98251e6ab71caea7))
* Change DYM decimals to 18 instead of 6 ([#240](https://github.com/dymensionxyz/roller/issues/240)) ([7db87bb](https://github.com/dymensionxyz/roller/commit/7db87bb7f127886ac23cd6ba5c47d20f61e90d90))
* change library installation flow ([#933](https://github.com/dymensionxyz/roller/issues/933)) ([97dedb2](https://github.com/dymensionxyz/roller/commit/97dedb265e6ef22cdac85001cd1d771eece7008d))
* change temp struct field type ([39d8620](https://github.com/dymensionxyz/roller/commit/39d8620963ab1ecfdf22d0649eaaa7a8fb4498d9))
* Change to use coin type 60 instead of 118 for the hub keys ([#543](https://github.com/dymensionxyz/roller/issues/543)) ([337db9f](https://github.com/dymensionxyz/roller/commit/337db9f61dae57ba5ffbeb0699f125b1b75b394c))
* changed default empty blocks to 90s ([#461](https://github.com/dymensionxyz/roller/issues/461)) ([98504f4](https://github.com/dymensionxyz/roller/commit/98504f44ad1e3f5bea6ce3c55040d8252020bf4d))
* check whether rollapp exists during initialization ([#829](https://github.com/dymensionxyz/roller/issues/829)) ([ded5e48](https://github.com/dymensionxyz/roller/commit/ded5e4848f7d157ecbdcbbe5f45290e355a3a2b7))
* cleanup unnecessary wallets, improve command output ([#854](https://github.com/dymensionxyz/roller/issues/854)) ([e9852f2](https://github.com/dymensionxyz/roller/commit/e9852f2df49ddff6173c97cf7ba80d92bdbf9642))
* config files ([#809](https://github.com/dymensionxyz/roller/issues/809)) ([aa7ff82](https://github.com/dymensionxyz/roller/commit/aa7ff82463710814dfa01409c2d8769836bbda45))
* **da:** don't ask to override the existing da config ([#999](https://github.com/dymensionxyz/roller/issues/999)) ([eb7e31c](https://github.com/dymensionxyz/roller/commit/eb7e31cb8f1a019766f674e592c97f2aee71566d))
* **deps:** bump eibc version ([#969](https://github.com/dymensionxyz/roller/issues/969)) ([6e7d0d5](https://github.com/dymensionxyz/roller/commit/6e7d0d57243280d0ac522ff4337ac353b662cbf1))
* **deps:** bump rollapp commit ([deb914b](https://github.com/dymensionxyz/roller/commit/deb914b2c95bd88d3e21e643df66cac667dc94db))
* **deps:** update dymd version to hub ([dd2dffd](https://github.com/dymensionxyz/roller/commit/dd2dffd96283bde4b1f556179296f6fc1717dc1a))
* **deps:** update eibc version ([48338c3](https://github.com/dymensionxyz/roller/commit/48338c300643cc59dc80c2090751f39c210a845f))
* **deps:** update rollapp commit and prebuilt version ([ca329c6](https://github.com/dymensionxyz/roller/commit/ca329c63a8bc768354d296b1a3abd361910b4a37))
* **deps:** update rollapp version ([639522c](https://github.com/dymensionxyz/roller/commit/639522c95e2eea806786dbb8b9f2d5b3dea4c3ea))
* **deps:** update rollapp version ([79c72db](https://github.com/dymensionxyz/roller/commit/79c72db28b482bf542537357a458e1f0d64a156a))
* Display balance on roller run in the bigger token denom ([#231](https://github.com/dymensionxyz/roller/issues/231)) ([04d6ffa](https://github.com/dymensionxyz/roller/commit/04d6fface8f840da43aaf678ecd8716d1620bbbb))
* display the relevant information based on node type ([8f82ffc](https://github.com/dymensionxyz/roller/commit/8f82ffc0db53bab38c186e29a9855e7828a40ed2))
* dymint expect mock in config instead of local ([#460](https://github.com/dymensionxyz/roller/issues/460)) ([2f3d444](https://github.com/dymensionxyz/roller/commit/2f3d4446a4333a41d29cce992130d3492c02ad6d))
* eibc client command error handling ([#837](https://github.com/dymensionxyz/roller/issues/837)) ([ed75b43](https://github.com/dymensionxyz/roller/commit/ed75b4379a1fbdac8eb63c16907ec6da1e41b5b6))
* eibc client commands ([#972](https://github.com/dymensionxyz/roller/issues/972)) ([3e650ae](https://github.com/dymensionxyz/roller/commit/3e650ae05fdf67d25f18afc3233ae984fc343cf5))
* **eibc:** improve error handling ([94f27a7](https://github.com/dymensionxyz/roller/commit/94f27a708c45718ed222b4a62ecdee6265d6f568))
* Enable Remote Connections by Exposing RollApp's JSON RPC Endpoint on 0.0.0.0 ([#275](https://github.com/dymensionxyz/roller/issues/275)) ([640d0f0](https://github.com/dymensionxyz/roller/commit/640d0f05e67d7012b83b23e90ad6aedf49e24753))
* Enhance error handling for 'roller register' common flows ([#53](https://github.com/dymensionxyz/roller/issues/53)) ([b0c360d](https://github.com/dymensionxyz/roller/commit/b0c360dbef289cf4662eeda3d1bf1c1b346b5b45))
* fail if the downloaded hash does not match the one registered with the rollapp ([286177d](https://github.com/dymensionxyz/roller/commit/286177d75ee583f4da26017f2c267c60cae73cc8))
* fail when snapshot checksum comparison doesn't pass ([bf45bd2](https://github.com/dymensionxyz/roller/commit/bf45bd271b909c1a554dcfead0dffdc9b7a0a36a))
* fix outputs that require node type ([4b218b8](https://github.com/dymensionxyz/roller/commit/4b218b858026c69558314c5d0395f5d762ea2ac0))
* Fix relayer start ([#320](https://github.com/dymensionxyz/roller/issues/320)) ([2cbbc6c](https://github.com/dymensionxyz/roller/commit/2cbbc6c6758200fd84a05a706dccb187c8df8564))
* Fix source dymd path in the CD flow ([#416](https://github.com/dymensionxyz/roller/issues/416)) ([ec6cd87](https://github.com/dymensionxyz/roller/commit/ec6cd877dcc0410a78bbf6b5101f1b17613e6c9c))
* Fixed connection-id missing key would cause relayer start to fail ([#503](https://github.com/dymensionxyz/roller/issues/503)) ([186c275](https://github.com/dymensionxyz/roller/commit/186c275d595ac367b9687410345c6f41e5bb6645))
* Fixed go version in build workflow ([#672](https://github.com/dymensionxyz/roller/issues/672)) ([fcef347](https://github.com/dymensionxyz/roller/commit/fcef347e8c44a77da014e64069f1d3e1faa26514))
* Fixed running `roller config init` when root dir does not exist ([#45](https://github.com/dymensionxyz/roller/issues/45)) ([81c8881](https://github.com/dymensionxyz/roller/commit/81c8881f06922d654111c1ec52c44f7013a065c6))
* Fixed spinner cursor removal bug ([#188](https://github.com/dymensionxyz/roller/issues/188)) ([cec5a9f](https://github.com/dymensionxyz/roller/commit/cec5a9fd219c4fdef011ce06cd4caca5b38b41fb))
* Generate formatted rollapp id for local rollapps ([#526](https://github.com/dymensionxyz/roller/issues/526)) ([b9fabcd](https://github.com/dymensionxyz/roller/commit/b9fabcd624444d7973a6e61fbcadb41ac2ec5b2e))
* handle &gt;1 sequencer when there are no state updates ([#839](https://github.com/dymensionxyz/roller/issues/839)) ([abc90b4](https://github.com/dymensionxyz/roller/commit/abc90b4d8ac6350acfa4fafcb4c75eb22a75adfe))
* handle `raw_log` error during transaction execution ([#922](https://github.com/dymensionxyz/roller/issues/922)) ([0280065](https://github.com/dymensionxyz/roller/commit/0280065366505e59439054b751e6f0d22a3fa4a1))
* handle empty genesis token supply ([#851](https://github.com/dymensionxyz/roller/issues/851)) ([ecb9135](https://github.com/dymensionxyz/roller/commit/ecb9135117a8891c8ad80316c6cbcf30e299c7ce))
* handle existing rollapp state for da sync ([c5f5189](https://github.com/dymensionxyz/roller/commit/c5f51893533728abf365f84f5c3a883382a2b90d))
* handle mock da in configuration update ([94c93d8](https://github.com/dymensionxyz/roller/commit/94c93d81d2a30c64e27cd6f610aab2834bc2ef54))
* **hub:** update playground hub chain id ([#962](https://github.com/dymensionxyz/roller/issues/962)) ([28851d5](https://github.com/dymensionxyz/roller/commit/28851d55f5546d59fa7cb63caf69e96cecceea48))
* ibc clients creation stuck  ([#639](https://github.com/dymensionxyz/roller/issues/639)) ([ff3983a](https://github.com/dymensionxyz/roller/commit/ff3983abe7acb4912323c7fefba126bb9c3f9cf6))
* improve mock settlement handling ([d040895](https://github.com/dymensionxyz/roller/commit/d040895a2c53fda6b7aec2d1c6c344a6bfec6768))
* improve relayer flow and address handling ([#869](https://github.com/dymensionxyz/roller/issues/869)) ([dcbe2d1](https://github.com/dymensionxyz/roller/commit/dcbe2d1406f8e3bcaef622dbe51c19b007edd703))
* install roller as part of binary installation ([#886](https://github.com/dymensionxyz/roller/issues/886)) ([32434d8](https://github.com/dymensionxyz/roller/commit/32434d83949623f62f36f6f3a30f4c74ca0a0152))
* **keys:** align key retrieval function with key creation ([#974](https://github.com/dymensionxyz/roller/issues/974)) ([d6ceb6c](https://github.com/dymensionxyz/roller/commit/d6ceb6ccc33a5d2b434b169031d772e929f3453a))
* **keys:** remove home from key directory where it's unnecessary ([168c871](https://github.com/dymensionxyz/roller/commit/168c871c2185fff39ced314b63d3edf8b031e694))
* lint ([6a43e45](https://github.com/dymensionxyz/roller/commit/6a43e458c31420afc66980bae709f310f3d9419d))
* linux binary installation for mock ([#917](https://github.com/dymensionxyz/roller/issues/917)) ([70e9d46](https://github.com/dymensionxyz/roller/commit/70e9d46498022cf13aa0d49fdc344cc448f7e08e))
* Make interactive flow exit on interrupt ([#228](https://github.com/dymensionxyz/roller/issues/228)) ([65ba75c](https://github.com/dymensionxyz/roller/commit/65ba75cc18d70cb5b9b5de59bb23f7aed8be7453))
* Make register return an error on insufficient funds ([#649](https://github.com/dymensionxyz/roller/issues/649)) ([d8537d4](https://github.com/dymensionxyz/roller/commit/d8537d4949e51ec8c3b8c7ddd82f7767162ddcfd))
* Match `roller config export` faucet URL with hub ID ([#537](https://github.com/dymensionxyz/roller/issues/537)) ([2811f63](https://github.com/dymensionxyz/roller/commit/2811f63031074e310cb93f71694f493ad24a42bd))
* Migrate panic when using mock DA ([#492](https://github.com/dymensionxyz/roller/issues/492)) ([e21833e](https://github.com/dymensionxyz/roller/commit/e21833e5e2556f1f901234c51e4d21de91c7c8ea))
* mock genesis generation ([#931](https://github.com/dymensionxyz/roller/issues/931)) ([4b08ea7](https://github.com/dymensionxyz/roller/commit/4b08ea7639a3ed458957962f6933e3e6c147744d))
* pass gas prices on roller register ([#253](https://github.com/dymensionxyz/roller/issues/253)) ([2022644](https://github.com/dymensionxyz/roller/commit/2022644ad63ea1644b3e6698f05b203b4cb47516))
* Prevent generate new avail seed in `roller migrate` ([#517](https://github.com/dymensionxyz/roller/issues/517)) ([ecfe396](https://github.com/dymensionxyz/roller/commit/ecfe39655edea3bd4b534907d4239e2bcce00389))
* pull docker image before running the container ([816d122](https://github.com/dymensionxyz/roller/commit/816d122b14185d4fef18b4aaebe4087237764b1d))
* query against a node when fetching bond params ([#857](https://github.com/dymensionxyz/roller/issues/857)) ([f562a6b](https://github.com/dymensionxyz/roller/commit/f562a6b09324dc2cf2bb418cc646164cddd944bf))
* Read output ports from config files ([#455](https://github.com/dymensionxyz/roller/issues/455)) ([38a57c4](https://github.com/dymensionxyz/roller/commit/38a57c4a70f8d6cc8417900f8cb9424e6eb7fc99))
* register show output ([#191](https://github.com/dymensionxyz/roller/issues/191)) ([df9ecac](https://github.com/dymensionxyz/roller/commit/df9ecacbe5ee2bb5cf7841040afb0e004383e2a5))
* relayer client creation flow ([#876](https://github.com/dymensionxyz/roller/issues/876)) ([2dddc4a](https://github.com/dymensionxyz/roller/commit/2dddc4a112ea22b92cc00799e726e30ad81c6da3))
* relayer ibc creation ([#935](https://github.com/dymensionxyz/roller/issues/935)) ([6b040f2](https://github.com/dymensionxyz/roller/commit/6b040f2b9c6b800efb2392f480b2a10e82b10ce0))
* **relayer run:** wait for rollapp to be healthy after block time change ([#861](https://github.com/dymensionxyz/roller/issues/861)) ([207b17a](https://github.com/dymensionxyz/roller/commit/207b17a025c0afb1283a4543ffe1a5d13eaff0ba))
* relayer setup flow for darwin ([#952](https://github.com/dymensionxyz/roller/issues/952)) ([d3c1705](https://github.com/dymensionxyz/roller/commit/d3c1705fff6cb354149a7ec99bd2e2b457ff4f33))
* relayer setup output should show the channels ([#944](https://github.com/dymensionxyz/roller/issues/944)) ([e5e4ac6](https://github.com/dymensionxyz/roller/commit/e5e4ac637b4465286af7e4d281b2b080385936d0))
* relayer won't print errors on the stdout, only in logfile ([#344](https://github.com/dymensionxyz/roller/issues/344)) ([ab81073](https://github.com/dymensionxyz/roller/commit/ab81073b734dae5fc10ceed8e13533db23090556))
* Reliable status for rollapp sequencer and relayer ([#204](https://github.com/dymensionxyz/roller/issues/204)) ([9ad3320](https://github.com/dymensionxyz/roller/commit/9ad3320b2d0e7e5378d26301bb521d2019f57181))
* remove systemd services on reset ([#899](https://github.com/dymensionxyz/roller/issues/899)) ([d00987b](https://github.com/dymensionxyz/roller/commit/d00987b8b0a2a4e33ccb6ee207ee5d8b8c2d3a20))
* remove trust-period override in relayer config ([#928](https://github.com/dymensionxyz/roller/issues/928)) ([5612d0e](https://github.com/dymensionxyz/roller/commit/5612d0e761d8b13b9c13db60cac27f0dc39ca312))
* Removed addresses that doesn't require funding from roller config init output ([#127](https://github.com/dymensionxyz/roller/issues/127)) ([d025248](https://github.com/dymensionxyz/roller/commit/d025248d57a8498b3c7d4b551fb95f7be1bd482d))
* removed hard coded override flag ([#395](https://github.com/dymensionxyz/roller/issues/395)) ([32446ad](https://github.com/dymensionxyz/roller/commit/32446ad77d071d199447ceff44fb987c8403de45))
* removed json-rpc flag from sequencer start ([#289](https://github.com/dymensionxyz/roller/issues/289)) ([a90cf87](https://github.com/dymensionxyz/roller/commit/a90cf87cd6b768f8026d8fc38f6de9bfda8ff1e7))
* Return relayer logs for periodic commands ([#558](https://github.com/dymensionxyz/roller/issues/558)) ([1e6decd](https://github.com/dymensionxyz/roller/commit/1e6decd5518f36153c9efd7cc7357017c75e695a))
* roller extarcts wrong active channel ([#498](https://github.com/dymensionxyz/roller/issues/498)) ([f0b0b17](https://github.com/dymensionxyz/roller/commit/f0b0b1739ed56709c0c536391ab6dc71ce3ac2c3))
* roller initialization ([#791](https://github.com/dymensionxyz/roller/issues/791)) ([7119edb](https://github.com/dymensionxyz/roller/commit/7119edbbbaf01a2f057b1e2ea3b8d11beeaaeea0))
* Roller run status services render always in the same order ([#206](https://github.com/dymensionxyz/roller/issues/206)) ([b338898](https://github.com/dymensionxyz/roller/commit/b33889853a2ba143aba40f5e69f538f68b8c807b))
* run da client on non-linux boxes with rollapp start ([#868](https://github.com/dymensionxyz/roller/issues/868)) ([79df251](https://github.com/dymensionxyz/roller/commit/79df251ef95bbd41d7418c956766b33478cb7358))
* set dymint settlement node using the correct key ([f121df4](https://github.com/dymensionxyz/roller/commit/f121df4f24778acd0914ab864a1b8c7b6ed1f07c))
* set max proof time to 1m ([#967](https://github.com/dymensionxyz/roller/issues/967)) ([5f33de1](https://github.com/dymensionxyz/roller/commit/5f33de1094161fa85e5463467a2661f64116580f))
* Set rollapp default listen addr to 0.0.0.0:26657 ([#345](https://github.com/dymensionxyz/roller/issues/345)) ([3661623](https://github.com/dymensionxyz/roller/commit/36616232a14fb902ac9544d3bf059d1177c21199))
* setting keyring-backend os for the client ([#283](https://github.com/dymensionxyz/roller/issues/283)) ([b084dc8](https://github.com/dymensionxyz/roller/commit/b084dc86391fdd17dcff999e749d77ae1c6205cc))
* show relayer log location ([#890](https://github.com/dymensionxyz/roller/issues/890)) ([eb99101](https://github.com/dymensionxyz/roller/commit/eb9910146f5d434988cec8937f227450fc966dd5))
* Switch from gvm to setup-go action in workflow ([#380](https://github.com/dymensionxyz/roller/issues/380)) ([74aabfb](https://github.com/dymensionxyz/roller/commit/74aabfb998c242a869e447cd813a551f0064214c))
* typos ([#570](https://github.com/dymensionxyz/roller/issues/570)) ([93d9bf2](https://github.com/dymensionxyz/roller/commit/93d9bf2d5c6f696cff0678efe73782cf13b183d0))
* **ui:** installation spinner shouldn't draw new lines ([62e25db](https://github.com/dymensionxyz/roller/commit/62e25dbe8b46981dec5ed50eaed7666663a4ad5c))
* Update `roller migrate` to Modify Rollapp Config for Zero Address RPC Exposure ([#348](https://github.com/dymensionxyz/roller/issues/348)) ([165332f](https://github.com/dymensionxyz/roller/commit/165332f4efc9374eb149cab1c5ddffaff36eb95e))
* update compile_locally rollapp-evm ver ([#436](https://github.com/dymensionxyz/roller/issues/436)) ([2debe7f](https://github.com/dymensionxyz/roller/commit/2debe7ff2b455cfb4d5a1df4006f2d6315ae496f))
* update compile_locally.sh with relayer version v0.1.6 ([#581](https://github.com/dymensionxyz/roller/issues/581)) ([c3fbc47](https://github.com/dymensionxyz/roller/commit/c3fbc4724b684aeae714439a047c587c46a0c32d))
* update create-sequencer command with a new arg sequence ([#859](https://github.com/dymensionxyz/roller/issues/859)) ([e8e0f2e](https://github.com/dymensionxyz/roller/commit/e8e0f2e17acdb66af05c3142485773f3549cf756))
* update relayer start cmd ([#883](https://github.com/dymensionxyz/roller/issues/883)) ([1d3cf4b](https://github.com/dymensionxyz/roller/commit/1d3cf4b740374355860241a824fe03248976ae34))
* Update rly path data instead of re-create ([#489](https://github.com/dymensionxyz/roller/issues/489)) ([df5d0fe](https://github.com/dymensionxyz/roller/commit/df5d0fed941095a2a18767a268fb167f883bfe3d))
* update status when running with active channel ([#522](https://github.com/dymensionxyz/roller/issues/522)) ([a066f8e](https://github.com/dymensionxyz/roller/commit/a066f8e7a0697f75ed07ef3a8427f9d9f59b960e))
* updated silknodes domain ([#214](https://github.com/dymensionxyz/roller/issues/214)) ([982e73c](https://github.com/dymensionxyz/roller/commit/982e73c8e75ae032ab2d2c43f243c37ce5d37f35))
* use `celestia-app` binary for balance q ([#860](https://github.com/dymensionxyz/roller/issues/860)) ([02d52fb](https://github.com/dymensionxyz/roller/commit/02d52fbe2237e8293f7064bbd236983420013f3b))
* use a chain ID from the available metadata ([#858](https://github.com/dymensionxyz/roller/issues/858)) ([f21a823](https://github.com/dymensionxyz/roller/commit/f21a823981076c0c2952607a4362e7e0a274bbb0))
* use adym as denom ([#740](https://github.com/dymensionxyz/roller/issues/740)) ([b4977ec](https://github.com/dymensionxyz/roller/commit/b4977ec4a6d9231fcd0c1ee1d869d5450665a75c))
* using evmos validation for rollappID ([#222](https://github.com/dymensionxyz/roller/issues/222)) ([97502af](https://github.com/dymensionxyz/roller/commit/97502afa1f07bd8f8830ecf384d435408042dbae))
* using os.Executable instead of hardcoded roller path ([#171](https://github.com/dymensionxyz/roller/issues/171)) ([c2e5518](https://github.com/dymensionxyz/roller/commit/c2e5518ac35def5fa7a46f2595b10de2728a92f7))
* **ux:** remove irrelevant prompts  ([#940](https://github.com/dymensionxyz/roller/issues/940)) ([74323ef](https://github.com/dymensionxyz/roller/commit/74323ef341ba7266f50a24458053d429e9145a32))
* **ux:** update prompts and `rollapp status` command ([#960](https://github.com/dymensionxyz/roller/issues/960)) ([c193839](https://github.com/dymensionxyz/roller/commit/c1938393f1fb9e92b7e5b068caa0dd306497facd))
* validate active connection ([#500](https://github.com/dymensionxyz/roller/issues/500)) ([c3fcb6f](https://github.com/dymensionxyz/roller/commit/c3fcb6fa6316295cb254424ccee3f40a37552c48))
* verify rollapp bech prefix against the build flag ([#877](https://github.com/dymensionxyz/roller/issues/877)) ([e036cc6](https://github.com/dymensionxyz/roller/commit/e036cc63f50773904a5e8bc48b4b7720d5b6cc13))
* Verify rollapp hub height is valid before creating a connection ([#467](https://github.com/dymensionxyz/roller/issues/467)) ([2d422e5](https://github.com/dymensionxyz/roller/commit/2d422e56030b1217cfe165fc8baebfc808f270d8))
* Verifying unique rollapp ID immediately in interactive flow  ([#226](https://github.com/dymensionxyz/roller/issues/226)) ([f033699](https://github.com/dymensionxyz/roller/commit/f03369964959946f6fbe6efd79999384fe516bd4))
* **wasm:** add wasm vm type ([#1000](https://github.com/dymensionxyz/roller/issues/1000)) ([ce08f6d](https://github.com/dymensionxyz/roller/commit/ce08f6dc3f36e250fe01b8466c7261851d47bc8c))
* Write new roller version to config after migration ([#491](https://github.com/dymensionxyz/roller/issues/491)) ([2a7016f](https://github.com/dymensionxyz/roller/commit/2a7016ff0730eeb4ae2186a44c940f7fc492fa94))
* writing denom metatada to the gensis file ([#631](https://github.com/dymensionxyz/roller/issues/631)) ([d0cf7bd](https://github.com/dymensionxyz/roller/commit/d0cf7bd91744a58ec424fc7e143639d3292d8a41))


### Miscellaneous Chores

* release 1.1.0-alpha-rc01 ([cb125c1](https://github.com/dymensionxyz/roller/commit/cb125c115bc429fb519c61a958011b4a0ffe6528))
* release 1.1.0-beta-rc01 ([0b956ac](https://github.com/dymensionxyz/roller/commit/0b956acc24e66959897f347ae1d4971b22f83874))
* release v1.7.0-alpha-pre-rc ([a4f07ac](https://github.com/dymensionxyz/roller/commit/a4f07ac7c3a34787d4a88f157e5657a11e83dd92))

## [1.6.4-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.6.3-alpha-rc01...v1.6.4-alpha-rc01) (2024-09-18)


### Bug Fixes

* **eibc:** improve error handling ([94f27a7](https://github.com/dymensionxyz/roller/commit/94f27a708c45718ed222b4a62ecdee6265d6f568))
* **keys:** remove home from key directory where it's unnecessary ([168c871](https://github.com/dymensionxyz/roller/commit/168c871c2185fff39ced314b63d3edf8b031e694))

## [1.6.3-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.6.2-alpha-rc01...v1.6.3-alpha-rc01) (2024-09-18)


### Bug Fixes

* **keys:** align key retrieval function with key creation ([#974](https://github.com/dymensionxyz/roller/issues/974)) ([d6ceb6c](https://github.com/dymensionxyz/roller/commit/d6ceb6ccc33a5d2b434b169031d772e929f3453a))

## [1.6.2-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.6.1-alpha-rc01...v1.6.2-alpha-rc01) (2024-09-18)


### Bug Fixes

* eibc client commands ([#972](https://github.com/dymensionxyz/roller/issues/972)) ([3e650ae](https://github.com/dymensionxyz/roller/commit/3e650ae05fdf67d25f18afc3233ae984fc343cf5))

## [1.6.1-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.6.0-alpha-rc01...v1.6.1-alpha-rc01) (2024-09-17)


### Bug Fixes

* **deps:** update eibc version ([48338c3](https://github.com/dymensionxyz/roller/commit/48338c300643cc59dc80c2090751f39c210a845f))

## [1.6.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.5.0-alpha-rc01...v1.6.0-alpha-rc01) (2024-09-17)


### Features

* add wasm rollapp support ([#968](https://github.com/dymensionxyz/roller/issues/968)) ([dc57755](https://github.com/dymensionxyz/roller/commit/dc57755a8574ec817cf40eba35d7647720974c79))


### Bug Fixes

* **deps:** bump eibc version ([#969](https://github.com/dymensionxyz/roller/issues/969)) ([6e7d0d5](https://github.com/dymensionxyz/roller/commit/6e7d0d57243280d0ac522ff4337ac353b662cbf1))

## [1.5.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.5-alpha-rc01...v1.5.0-alpha-rc01) (2024-09-17)


### Features

* add progress bar to downloads ([#965](https://github.com/dymensionxyz/roller/issues/965)) ([e6d3b4b](https://github.com/dymensionxyz/roller/commit/e6d3b4b8ce11783e55a96b9e85a0c41414cc2954))


### Bug Fixes

* set max proof time to 1m ([#967](https://github.com/dymensionxyz/roller/issues/967)) ([5f33de1](https://github.com/dymensionxyz/roller/commit/5f33de1094161fa85e5463467a2661f64116580f))

## [1.4.5-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.4-alpha-rc01...v1.4.5-alpha-rc01) (2024-09-17)


### Bug Fixes

* **deps:** update rollapp commit and prebuilt version ([ca329c6](https://github.com/dymensionxyz/roller/commit/ca329c63a8bc768354d296b1a3abd361910b4a37))

## [1.4.4-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.3-alpha-rc01...v1.4.4-alpha-rc01) (2024-09-17)


### Bug Fixes

* **hub:** update playground hub chain id ([#962](https://github.com/dymensionxyz/roller/issues/962)) ([28851d5](https://github.com/dymensionxyz/roller/commit/28851d55f5546d59fa7cb63caf69e96cecceea48))

## [1.4.3-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.2-alpha-rc01...v1.4.3-alpha-rc01) (2024-09-17)


### Bug Fixes

* **deps:** update dymd version to hub ([dd2dffd](https://github.com/dymensionxyz/roller/commit/dd2dffd96283bde4b1f556179296f6fc1717dc1a))
* **ux:** update prompts and `rollapp status` command ([#960](https://github.com/dymensionxyz/roller/issues/960)) ([c193839](https://github.com/dymensionxyz/roller/commit/c1938393f1fb9e92b7e5b068caa0dd306497facd))

## [1.4.2-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.1-alpha-rc01...v1.4.2-alpha-rc01) (2024-09-16)


### Bug Fixes

* **deps:** bump rollapp commit ([deb914b](https://github.com/dymensionxyz/roller/commit/deb914b2c95bd88d3e21e643df66cac667dc94db))

## [1.4.1-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.4.0-alpha-rc01...v1.4.1-alpha-rc01) (2024-09-16)


### Bug Fixes

* **deps:** update rollapp version ([639522c](https://github.com/dymensionxyz/roller/commit/639522c95e2eea806786dbb8b9f2d5b3dea4c3ea))
* **deps:** update rollapp version ([79c72db](https://github.com/dymensionxyz/roller/commit/79c72db28b482bf542537357a458e1f0d64a156a))

## [1.4.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.3.0-alpha-rc01...v1.4.0-alpha-rc01) (2024-09-16)


### Features

* add cron job for relayer flush ([#954](https://github.com/dymensionxyz/roller/issues/954)) ([24d39dd](https://github.com/dymensionxyz/roller/commit/24d39ddecdff423ec6084899f83eb12905a437c2))


### Bug Fixes

* relayer setup flow for darwin ([#952](https://github.com/dymensionxyz/roller/issues/952)) ([d3c1705](https://github.com/dymensionxyz/roller/commit/d3c1705fff6cb354149a7ec99bd2e2b457ff4f33))

## [1.3.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.2.4-alpha-rc01...v1.3.0-alpha-rc01) (2024-09-16)


### Features

* add launchctl service support for macos ([#951](https://github.com/dymensionxyz/roller/issues/951)) ([f518849](https://github.com/dymensionxyz/roller/commit/f5188497c9b6abd489a3d72289b4e2fd2c0e463a))
* add logs command for services ([#943](https://github.com/dymensionxyz/roller/issues/943)) ([360644f](https://github.com/dymensionxyz/roller/commit/360644f64bc65ae65c44d64c8a63cf95f52fbefa))
* add sequencer address and balance to `rollapp` output commands ([#948](https://github.com/dymensionxyz/roller/issues/948)) ([7668320](https://github.com/dymensionxyz/roller/commit/7668320ca951942403baa9dda5204e3751a1738f))


### Bug Fixes

* relayer setup output should show the channels ([#944](https://github.com/dymensionxyz/roller/issues/944)) ([e5e4ac6](https://github.com/dymensionxyz/roller/commit/e5e4ac637b4465286af7e4d281b2b080385936d0))
* **ux:** remove irrelevant prompts  ([#940](https://github.com/dymensionxyz/roller/issues/940)) ([74323ef](https://github.com/dymensionxyz/roller/commit/74323ef341ba7266f50a24458053d429e9145a32))

## [1.2.4-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.2.3-alpha-rc01...v1.2.4-alpha-rc01) (2024-09-13)


### Bug Fixes

* relayer ibc creation ([#935](https://github.com/dymensionxyz/roller/issues/935)) ([6b040f2](https://github.com/dymensionxyz/roller/commit/6b040f2b9c6b800efb2392f480b2a10e82b10ce0))

## [1.2.3-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.2.2-alpha-rc01...v1.2.3-alpha-rc01) (2024-09-13)


### Bug Fixes

* change library installation flow ([#933](https://github.com/dymensionxyz/roller/issues/933)) ([97dedb2](https://github.com/dymensionxyz/roller/commit/97dedb265e6ef22cdac85001cd1d771eece7008d))

## [1.2.2-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.2.1-alpha-rc01...v1.2.2-alpha-rc01) (2024-09-13)


### Bug Fixes

* mock genesis generation ([#931](https://github.com/dymensionxyz/roller/issues/931)) ([4b08ea7](https://github.com/dymensionxyz/roller/commit/4b08ea7639a3ed458957962f6933e3e6c147744d))
* remove trust-period override in relayer config ([#928](https://github.com/dymensionxyz/roller/issues/928)) ([5612d0e](https://github.com/dymensionxyz/roller/commit/5612d0e761d8b13b9c13db60cac27f0dc39ca312))

## [1.2.1-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.2.0-alpha-rc01...v1.2.1-alpha-rc01) (2024-09-13)


### Bug Fixes

* handle `raw_log` error during transaction execution ([#922](https://github.com/dymensionxyz/roller/issues/922)) ([0280065](https://github.com/dymensionxyz/roller/commit/0280065366505e59439054b751e6f0d22a3fa4a1))

## [1.2.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.1.0-alpha-rc01...v1.2.0-alpha-rc01) (2024-09-12)


### Features

* make it possible to fill out the moniker and social media links for sequencer ([#923](https://github.com/dymensionxyz/roller/issues/923)) ([a8cf922](https://github.com/dymensionxyz/roller/commit/a8cf9221e5b08c85700b58368f46f16d7dbdc967))
* validate rollapp ID before dependency installation ([#919](https://github.com/dymensionxyz/roller/issues/919)) ([bcf59eb](https://github.com/dymensionxyz/roller/commit/bcf59eb987bad9f280a97e46217e1ceee94e70d7))


### Bug Fixes

* change default confirm value to false for all interactions ([#921](https://github.com/dymensionxyz/roller/issues/921)) ([c5d3912](https://github.com/dymensionxyz/roller/commit/c5d391207d82190d21af71ef98251e6ab71caea7))

## [1.1.0-alpha-rc01](https://github.com/dymensionxyz/roller/compare/v1.0.1-beta...v1.1.0-alpha-rc01) (2024-09-11)


### Features

* add `--mock` flag to init ([#912](https://github.com/dymensionxyz/roller/issues/912)) ([0fccdb3](https://github.com/dymensionxyz/roller/commit/0fccdb3cb4b15b69ead383bad6c20493ab1f4f82))
* add binary installation command ([#879](https://github.com/dymensionxyz/roller/issues/879)) ([71657bd](https://github.com/dymensionxyz/roller/commit/71657bd8a5cb23196968aa683b0e3c88e56fc5b4))
* add block explorer command ([#880](https://github.com/dymensionxyz/roller/issues/880)) ([73e0cb2](https://github.com/dymensionxyz/roller/commit/73e0cb2845014b8f812efc238a3d40ef8dc5406b))
* add commands to export and update sequencer metadata ([#847](https://github.com/dymensionxyz/roller/issues/847)) ([3febe76](https://github.com/dymensionxyz/roller/commit/3febe76fea22c2afa994c113a790d57a92597db4))
* add eibc client support ([#813](https://github.com/dymensionxyz/roller/issues/813)) ([89dc1c6](https://github.com/dymensionxyz/roller/commit/89dc1c67bcb54510f44bf17eeb4cfd634dd25a7d))
* add rollapp init ([#803](https://github.com/dymensionxyz/roller/issues/803)) ([19a455a](https://github.com/dymensionxyz/roller/commit/19a455a0d30c58b424408b80741ba7f2a21cf027))
* add rollapp status command ([#812](https://github.com/dymensionxyz/roller/issues/812)) ([0e1d2a1](https://github.com/dymensionxyz/roller/commit/0e1d2a1ee7cb17f5586e81693ebdc0643106a475))
* celestia light client config logic, sync from da head for first sequencer, sync from first state update for others ([#832](https://github.com/dymensionxyz/roller/issues/832)) ([c911b45](https://github.com/dymensionxyz/roller/commit/c911b451d90c468a739692e2d182fd2023694669))
* display and accept bond amount in base denom ([#911](https://github.com/dymensionxyz/roller/issues/911)) ([61b445b](https://github.com/dymensionxyz/roller/commit/61b445b210f51ebb08ad9917a2805e7528f30648))
* fetch DA information from `roller.toml` for easy endpoint changes ([#862](https://github.com/dymensionxyz/roller/issues/862)) ([1d66c64](https://github.com/dymensionxyz/roller/commit/1d66c64d1b7e3c895fb7f465eadee5fb4f4f9253))
* fetch rollapp info from chain ([#828](https://github.com/dymensionxyz/roller/issues/828)) ([130eb89](https://github.com/dymensionxyz/roller/commit/130eb8997da75559afd1255048b00f86ba7a9473))
* finalize full node flow ([#836](https://github.com/dymensionxyz/roller/issues/836)) ([ae3000b](https://github.com/dymensionxyz/roller/commit/ae3000bb6d42854181274e5c7263627eee48931b))
* finalize relayer flow ([#833](https://github.com/dymensionxyz/roller/issues/833)) ([ef430b1](https://github.com/dymensionxyz/roller/commit/ef430b1b460c8858d65271498bf36581355d3583))
* finalize sequencer registration flow ([#831](https://github.com/dymensionxyz/roller/issues/831)) ([3c0296a](https://github.com/dymensionxyz/roller/commit/3c0296a3b5c807a6fcb7665c9f32f0e3c9369eaa))
* full node flow ([#827](https://github.com/dymensionxyz/roller/issues/827)) ([95a9bff](https://github.com/dymensionxyz/roller/commit/95a9bff4263efdf57f5a471022ceee81598e8a98))
* generate chain config during block-explorer initialization ([#884](https://github.com/dymensionxyz/roller/issues/884)) ([c8afc46](https://github.com/dymensionxyz/roller/commit/c8afc46267670485a579be336a99b4e7f88520de))
* handle genesis creator archive ([#800](https://github.com/dymensionxyz/roller/issues/800)) ([e3ea97c](https://github.com/dymensionxyz/roller/commit/e3ea97c123a68ffa7937199fc81876b25dfc6494))
* handle sequencer bond ([#906](https://github.com/dymensionxyz/roller/issues/906)) ([c18e989](https://github.com/dymensionxyz/roller/commit/c18e9890b8a7f9c58299b420e0bc3f3915df24ac))
* handle sequencer reward address ([#866](https://github.com/dymensionxyz/roller/issues/866)) ([87f3765](https://github.com/dymensionxyz/roller/commit/87f37659c1e64e6697039358a7c602d031230615))
* improve rollapp initialization ([#817](https://github.com/dymensionxyz/roller/issues/817)) ([d7db004](https://github.com/dymensionxyz/roller/commit/d7db0042ba18d342f92ed160fd8f18659dbb7186))
* install binaries based on rollapp id ([#885](https://github.com/dymensionxyz/roller/issues/885)) ([ce93e8d](https://github.com/dymensionxyz/roller/commit/ce93e8dca399335157d808327bb2764e4cec2563))
* run relayer using roller ([#819](https://github.com/dymensionxyz/roller/issues/819)) ([cf84f05](https://github.com/dymensionxyz/roller/commit/cf84f055e0635abeb349dcb396f6315e6f608a90))
* separate service command into rollapp specific services ([3a8c36f](https://github.com/dymensionxyz/roller/commit/3a8c36f7f74dae64e6a3dbe95abd902eb10aecd6))
* sequencer registration ([#821](https://github.com/dymensionxyz/roller/issues/821)) ([8e7fe60](https://github.com/dymensionxyz/roller/commit/8e7fe60b52d03614c2c206518caec70921814fc6))
* support sequencer bond for registration ([#762](https://github.com/dymensionxyz/roller/issues/762)) ([1ca9f55](https://github.com/dymensionxyz/roller/commit/1ca9f559f673fcdcfce29762696ae4f60722e79d))
* update relayer flow to support lc ([#875](https://github.com/dymensionxyz/roller/issues/875)) ([4d494d5](https://github.com/dymensionxyz/roller/commit/4d494d5915b832b4aa6bb8346c0fe7a48038b65f))
* update sequencer metadata ([#850](https://github.com/dymensionxyz/roller/issues/850)) ([2166bc6](https://github.com/dymensionxyz/roller/commit/2166bc6095736a8343b6bdb161dd66202e03d114))
* wrap eibc version with min fee percentage support ([#874](https://github.com/dymensionxyz/roller/issues/874)) ([ae9432f](https://github.com/dymensionxyz/roller/commit/ae9432faa332aba64834f8aeae31d5d6f1480124))


### Bug Fixes

* add genesis account to genesis file ([#815](https://github.com/dymensionxyz/roller/issues/815)) ([2eb6a09](https://github.com/dymensionxyz/roller/commit/2eb6a090d4438e2bf1a33477866a2dfcf7d0f619))
* add sequencer addresses to the generated genesis file ([#808](https://github.com/dymensionxyz/roller/issues/808)) ([01827f4](https://github.com/dymensionxyz/roller/commit/01827f4bcdb566adb8b3710a2b23ccc96b101331))
* celestia namespace retrieval ([a0dd6f6](https://github.com/dymensionxyz/roller/commit/a0dd6f6f6efce445a109e71380dbc272d4b91666))
* change temp struct field type ([39d8620](https://github.com/dymensionxyz/roller/commit/39d8620963ab1ecfdf22d0649eaaa7a8fb4498d9))
* check whether rollapp exists during initialization ([#829](https://github.com/dymensionxyz/roller/issues/829)) ([ded5e48](https://github.com/dymensionxyz/roller/commit/ded5e4848f7d157ecbdcbbe5f45290e355a3a2b7))
* cleanup unnecessary wallets, improve command output ([#854](https://github.com/dymensionxyz/roller/issues/854)) ([e9852f2](https://github.com/dymensionxyz/roller/commit/e9852f2df49ddff6173c97cf7ba80d92bdbf9642))
* config files ([#809](https://github.com/dymensionxyz/roller/issues/809)) ([aa7ff82](https://github.com/dymensionxyz/roller/commit/aa7ff82463710814dfa01409c2d8769836bbda45))
* eibc client command error handling ([#837](https://github.com/dymensionxyz/roller/issues/837)) ([ed75b43](https://github.com/dymensionxyz/roller/commit/ed75b4379a1fbdac8eb63c16907ec6da1e41b5b6))
* fail if the downloaded hash does not match the one registered with the rollapp ([286177d](https://github.com/dymensionxyz/roller/commit/286177d75ee583f4da26017f2c267c60cae73cc8))
* fail when snapshot checksum comparison doesn't pass ([bf45bd2](https://github.com/dymensionxyz/roller/commit/bf45bd271b909c1a554dcfead0dffdc9b7a0a36a))
* Fixed go version in build workflow ([#672](https://github.com/dymensionxyz/roller/issues/672)) ([fcef347](https://github.com/dymensionxyz/roller/commit/fcef347e8c44a77da014e64069f1d3e1faa26514))
* handle &gt;1 sequencer when there are no state updates ([#839](https://github.com/dymensionxyz/roller/issues/839)) ([abc90b4](https://github.com/dymensionxyz/roller/commit/abc90b4d8ac6350acfa4fafcb4c75eb22a75adfe))
* handle empty genesis token supply ([#851](https://github.com/dymensionxyz/roller/issues/851)) ([ecb9135](https://github.com/dymensionxyz/roller/commit/ecb9135117a8891c8ad80316c6cbcf30e299c7ce))
* handle existing rollapp state for da sync ([c5f5189](https://github.com/dymensionxyz/roller/commit/c5f51893533728abf365f84f5c3a883382a2b90d))
* handle mock da in configuration update ([94c93d8](https://github.com/dymensionxyz/roller/commit/94c93d81d2a30c64e27cd6f610aab2834bc2ef54))
* ibc clients creation stuck  ([#639](https://github.com/dymensionxyz/roller/issues/639)) ([ff3983a](https://github.com/dymensionxyz/roller/commit/ff3983abe7acb4912323c7fefba126bb9c3f9cf6))
* improve mock settlement handling ([d040895](https://github.com/dymensionxyz/roller/commit/d040895a2c53fda6b7aec2d1c6c344a6bfec6768))
* improve relayer flow and address handling ([#869](https://github.com/dymensionxyz/roller/issues/869)) ([dcbe2d1](https://github.com/dymensionxyz/roller/commit/dcbe2d1406f8e3bcaef622dbe51c19b007edd703))
* install roller as part of binary installation ([#886](https://github.com/dymensionxyz/roller/issues/886)) ([32434d8](https://github.com/dymensionxyz/roller/commit/32434d83949623f62f36f6f3a30f4c74ca0a0152))
* lint ([6a43e45](https://github.com/dymensionxyz/roller/commit/6a43e458c31420afc66980bae709f310f3d9419d))
* linux binary installation for mock ([#917](https://github.com/dymensionxyz/roller/issues/917)) ([70e9d46](https://github.com/dymensionxyz/roller/commit/70e9d46498022cf13aa0d49fdc344cc448f7e08e))
* Make register return an error on insufficient funds ([#649](https://github.com/dymensionxyz/roller/issues/649)) ([d8537d4](https://github.com/dymensionxyz/roller/commit/d8537d4949e51ec8c3b8c7ddd82f7767162ddcfd))
* pull docker image before running the container ([816d122](https://github.com/dymensionxyz/roller/commit/816d122b14185d4fef18b4aaebe4087237764b1d))
* query against a node when fetching bond params ([#857](https://github.com/dymensionxyz/roller/issues/857)) ([f562a6b](https://github.com/dymensionxyz/roller/commit/f562a6b09324dc2cf2bb418cc646164cddd944bf))
* relayer client creation flow ([#876](https://github.com/dymensionxyz/roller/issues/876)) ([2dddc4a](https://github.com/dymensionxyz/roller/commit/2dddc4a112ea22b92cc00799e726e30ad81c6da3))
* **relayer run:** wait for rollapp to be healthy after block time change ([#861](https://github.com/dymensionxyz/roller/issues/861)) ([207b17a](https://github.com/dymensionxyz/roller/commit/207b17a025c0afb1283a4543ffe1a5d13eaff0ba))
* remove systemd services on reset ([#899](https://github.com/dymensionxyz/roller/issues/899)) ([d00987b](https://github.com/dymensionxyz/roller/commit/d00987b8b0a2a4e33ccb6ee207ee5d8b8c2d3a20))
* roller initialization ([#791](https://github.com/dymensionxyz/roller/issues/791)) ([7119edb](https://github.com/dymensionxyz/roller/commit/7119edbbbaf01a2f057b1e2ea3b8d11beeaaeea0))
* run da client on non-linux boxes with rollapp start ([#868](https://github.com/dymensionxyz/roller/issues/868)) ([79df251](https://github.com/dymensionxyz/roller/commit/79df251ef95bbd41d7418c956766b33478cb7358))
* set dymint settlement node using the correct key ([f121df4](https://github.com/dymensionxyz/roller/commit/f121df4f24778acd0914ab864a1b8c7b6ed1f07c))
* show relayer log location ([#890](https://github.com/dymensionxyz/roller/issues/890)) ([eb99101](https://github.com/dymensionxyz/roller/commit/eb9910146f5d434988cec8937f227450fc966dd5))
* update create-sequencer command with a new arg sequence ([#859](https://github.com/dymensionxyz/roller/issues/859)) ([e8e0f2e](https://github.com/dymensionxyz/roller/commit/e8e0f2e17acdb66af05c3142485773f3549cf756))
* update relayer start cmd ([#883](https://github.com/dymensionxyz/roller/issues/883)) ([1d3cf4b](https://github.com/dymensionxyz/roller/commit/1d3cf4b740374355860241a824fe03248976ae34))
* use `celestia-app` binary for balance q ([#860](https://github.com/dymensionxyz/roller/issues/860)) ([02d52fb](https://github.com/dymensionxyz/roller/commit/02d52fbe2237e8293f7064bbd236983420013f3b))
* use a chain ID from the available metadata ([#858](https://github.com/dymensionxyz/roller/issues/858)) ([f21a823](https://github.com/dymensionxyz/roller/commit/f21a823981076c0c2952607a4362e7e0a274bbb0))
* use adym as denom ([#740](https://github.com/dymensionxyz/roller/issues/740)) ([b4977ec](https://github.com/dymensionxyz/roller/commit/b4977ec4a6d9231fcd0c1ee1d869d5450665a75c))
* verify rollapp bech prefix against the build flag ([#877](https://github.com/dymensionxyz/roller/issues/877)) ([e036cc6](https://github.com/dymensionxyz/roller/commit/e036cc63f50773904a5e8bc48b4b7720d5b6cc13))


### Miscellaneous

* release 1.1.0-alpha-rc01 ([cb125c1](https://github.com/dymensionxyz/roller/commit/cb125c115bc429fb519c61a958011b4a0ffe6528))
* release 1.1.0-beta-rc01 ([0b956ac](https://github.com/dymensionxyz/roller/commit/0b956acc24e66959897f347ae1d4971b22f83874))

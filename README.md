![ci](https://github.com/six78/2-story-points-cli/actions/workflows/ci.yml/badge.svg)
[![Maintainability](https://api.codeclimate.com/v1/badges/7159536b897586bb0137/maintainability)](https://codeclimate.com/github/six78/2-story-points-cli/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/7159536b897586bb0137/test_coverage)](https://codeclimate.com/github/six78/2-story-points-cli/test_coverage)

# Lock, Stock and Two Story Points

### Decentralized poker planning in terminal

<p align="left">
  <img height="450" src="docs/demo.gif">
</p>

[//]: # (Fancy a web version? -> https://six78.github.io/2-story-points )

# Description

- This is a CLI app for poker planning
- We use [Waku](https://waku.org) for decantralized players communication
- Messages are end-to-end encrypted, the key is shared elsewhere as part of the room id

[//]: # (# Get it)

# Build it your own
 ```shell
 git clone https://github.io/six78/2-story-points-cli
 cd 2-story-points
 make build
 ./2sp
 ```
 
Or just run the code with a shadow build:

```shell
make run
```

Now share your room id with friends and start estimating your issues!

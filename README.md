TFGCo Go Extensions
===================

[![Build Status](https://travis-ci.org/topfreegames/extensions.svg?branch=master)](https://travis-ci.org/topfreegames/extensions)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/extensions/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/extensions?branch=master)

### UPDATE: Current Status
We've separated helpers which we judged as core into individual modules. In case you want to use one of them:
* [http](https://github.com/topfreegames/go-extensions-http)
* [tracing](https://github.com/topfreegames/go-extensions-tracing)
* [mongo](https://github.com/topfreegames/go-extensions-mongo)
* [s3](https://github.com/topfreegames/go-extensions-s3)
* [redis](https://github.com/topfreegames/go-extensions-redis)
* [kafka](https://github.com/topfreegames/go-extensions-kafka)
* [k8s-client-go](https://github.com/topfreegames/go-extensions-k8s-client-go)

This package contains the common extensions we use in our projects.

### Extensions
* Postgres
* Statsd
* Kafka Consumer
* Kafka Producer
* Redis

### Dependencias
* [librdkafka](https://github.com/edenhill/librdkafka)

### Changelog
#### v1.0.0

New Extensions:

* Posgres
* Statsd
* Kafka consumer
* Kafka producer

#### v1.1.0

* Dep support.

#### v1.2.0

New Extension:

* Redis

#### v1.2.1

Bugfixes:

* PG extension bugfix

#### v2.0.0

Breaking Changes:

* Each extension has its own package now

### v9.0.0

Breaking Changes:

* Depreciate support to go deps
* Add go modules `github.com/topfreegames/extensions/v9`
* Update minimum go version to 1.12 (including travis)
* The `jaeger-client-go` requires a dependency that change its name. To keep compability replace the required name by your module while the community don't release a new version with this fixed:
```
$ go mod edit -replace github.com/codahale/hdrhistogram=github.com/HdrHistogram/hdrhistogram-go@v0.0.0-20200919145931-8dac23c8dac1
```

#### v9.1.0

* Add support to Redis v8 (`github.com/go-redis/redis/v8`)

#### v9.2.0

* Add gRPC client tracing methods
* Fix github.com/codahale/hdrhistogram dependency import issue

#### v9.3.0

* Fix postgresql transaction orm operations.


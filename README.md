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

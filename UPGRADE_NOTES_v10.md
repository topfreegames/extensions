# Upgrade Notes for Extensions v10

This document outlines breaking changes and important considerations when upgrading to extensions v10.

## Table of Contents

- [Overview](#overview)
- [Redis](#redis)
- [PostgreSQL (go-pg)](#postgresql-go-pg)
- [MongoDB](#mongodb)
- [Cassandra](#cassandra)
- [AWS S3](#aws-s3)
- [Kafka](#kafka)
- [StatsD](#statsd)
- [Echo Web Framework](#echo-web-framework)
- [Testing (Ginkgo/Gomega)](#testing-ginkgogomega)
- [Mocking (gomock)](#mocking-gomock)
- [Other Libraries](#other-libraries)

---

## Overview

Extensions v10 includes major upgrades to most underlying libraries. This document highlights breaking changes that may affect your application.

**Minimum Go Version**: Go 1.24.0

---

## Redis

**Upgraded**: `github.com/go-redis/redis/v8` → `github.com/redis/go-redis/v9`

### Breaking Changes

#### 1. Import Path Changed
```diff
- import "github.com/go-redis/redis/v8"
+ import "github.com/redis/go-redis/v9"
```

#### 2. All Methods Now Require Context
All Redis commands now require a `context.Context` as the first parameter:

```diff
- client.Ping()
+ client.Ping(ctx)

- client.Get("key")
+ client.Get(ctx, "key")

- client.Set("key", "value", 0)
+ client.Set(ctx, "key", "value", 0)
```

#### 3. Context Methods Removed
The `Context()` and `WithContext()` methods have been removed. Context is now passed per-command:

```diff
- redis.WithContext(ctx).Get("key")
+ redis.Get(ctx, "key")
```

#### 4. Type Name Changes
- `StringStringMapCmd` → `MapStringStringCmd`
- `HMSet` now returns `*BoolCmd` instead of `*StatusCmd`

#### 5. Lock Library Changed
**Changed**: `github.com/bsm/redis-lock` → `github.com/bsm/redislock v0.9.4`

The lock API has changed:
```diff
- import "github.com/bsm/redis-lock"
+ import "github.com/bsm/redislock"

- lock.IsLocked() // Method removed
```

#### 6. Instrumentation Changes
The tracing instrumentation now uses hooks instead of middleware:
- `DialHook`
- `ProcessHook`
- `ProcessPipelineHook`

### Migration Guide

1. Add `context.Context` to all Redis calls
2. Remove usage of `Context()` and `WithContext()` methods
3. Update lock implementations (no more `IsLocked()` method)
4. Update any custom instrumentation to use the new hook system

---

## PostgreSQL (go-pg)

**Upgraded**: `github.com/go-pg/pg` v6 → `github.com/go-pg/pg/v10`

### Breaking Changes

#### 1. Import Path Changed
```diff
- import "github.com/go-pg/pg"
- import "github.com/go-pg/pg/orm"
+ import "github.com/go-pg/pg/v10"
+ import "github.com/go-pg/pg/v10/orm"
```

#### 2. ORM Method Signatures Changed
CRUD methods now return `(orm.Result, error)` instead of just `error`:

```diff
- Insert(model ...interface{}) error
+ Insert(model ...interface{}) (orm.Result, error)

- Update(model interface{}) error
+ Update(model interface{}) (orm.Result, error)

- Delete(model interface{}) error
+ Delete(model interface{}) (orm.Result, error)
```

**Migration example:**
```diff
- err := db.Insert(&user)
+ result, err := db.Insert(&user)
+ if err != nil {
+     return err
+ }
+ rowsAffected := result.RowsAffected()
```

#### 3. Context-Aware Methods Added
New context-aware methods have been added:
- `ExecContext(ctx, query, params...)`
- `ExecOneContext(ctx, query, params...)`
- `QueryContext(ctx, model, query, params...)`
- `QueryOneContext(ctx, model, query, params...)`

#### 4. Model() Required for Queries
In v10, you must call `Model()` before executing operations:

```diff
- db.Select(&users)
+ db.Model(&users).Select()
```

#### 5. Query Formatting Changed
`AppendQuery` now requires a formatter as the first parameter:

```diff
- val.AppendQuery(nil)
+ val.AppendQuery(orm.NewFormatter(), nil)
```

### Migration Guide

1. Update all import paths to `/v10`
2. Handle `orm.Result` return value from CRUD operations
3. Add `Model()` calls before query operations
4. Consider using new context-aware methods for better cancellation support

### Security Note

⚠️ **Important**: v10 fixes several SQL injection vulnerabilities present in earlier versions. Upgrading is strongly recommended for security.

---

## MongoDB

**Upgraded**: `gopkg.in/mgo.v2` → `go.mongodb.org/mongo-driver/v2`

### Breaking Changes

This is a **major rewrite** from the unmaintained `mgo` library to the official MongoDB Go driver.

#### 1. Import Path Completely Changed
```diff
- import "gopkg.in/mgo.v2"
- import "gopkg.in/mgo.v2/bson"
+ import "go.mongodb.org/mongo-driver/v2/mongo"
+ import "go.mongodb.org/mongo-driver/v2/bson"
+ import "go.mongodb.org/mongo-driver/v2/options"
```

#### 2. API Completely Redesigned

The new driver has a completely different API. Key changes:

**Connection:**
```diff
- session, err := mgo.Dial(url)
+ client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
```

**Database/Collection Access:**
```diff
- db := session.DB("mydb")
- collection := db.C("mycollection")
+ db := client.Database("mydb")
+ collection := db.Collection("mycollection")
```

**Find Operations:**
```diff
- err := collection.Find(bson.M{"name": "John"}).One(&result)
+ err := collection.FindOne(ctx, bson.M{"name": "John"}).Decode(&result)

- iter := collection.Find(query).Iter()
+ cursor, err := collection.Find(ctx, query)
+ defer cursor.Close(ctx)
+ for cursor.Next(ctx) {
+     cursor.Decode(&result)
+ }
```

**Insert Operations:**
```diff
- err := collection.Insert(doc)
+ result, err := collection.InsertOne(ctx, doc)

- err := collection.Insert(doc1, doc2)
+ result, err := collection.InsertMany(ctx, []interface{}{doc1, doc2})
```

**Update Operations:**
```diff
- err := collection.Update(selector, update)
+ result, err := collection.UpdateOne(ctx, selector, update)

- err := collection.UpdateId(id, update)
+ result, err := collection.UpdateByID(ctx, id, update)
```

**Delete Operations:**
```diff
- err := collection.Remove(selector)
+ result, err := collection.DeleteOne(ctx, selector)

- err := collection.RemoveAll(selector)
+ result, err := collection.DeleteMany(ctx, selector)
```

#### 3. All Methods Now Require Context
Every database operation requires a `context.Context` as the first parameter.

#### 4. BSON Types Changed
```diff
- bson.M, bson.D (from mgo)
+ bson.M, bson.D (from mongo-driver, similar but not identical)

- bson.ObjectId
+ primitive.ObjectID
```

#### 5. Session Handling Removed
The concept of sessions has changed. The driver manages connection pooling internally.

### Migration Guide

1. This is effectively a complete rewrite
2. Review all MongoDB operations in your codebase
3. Test thoroughly - the APIs are fundamentally different
4. Update all BSON type references
5. Add context to all database operations
6. Consider using the new aggregation pipeline features

---

## Cassandra

**Upgraded**: `github.com/gocql/gocql` → `github.com/apache/cassandra-gocql-driver/v2`

### Breaking Changes

#### 1. Import Path Changed
```diff
- import "github.com/gocql/gocql"
+ import gocql "github.com/apache/cassandra-gocql-driver/v2"
```

#### 2. NewCluster Signature Changed
`NewCluster()` now requires at least one host argument:

```diff
- cluster := gocql.NewCluster()
- cluster.Hosts = []string{"localhost"}
+ cluster := gocql.NewCluster("localhost")
```

### Migration Guide

1. Update import paths
2. Pass hosts directly to `NewCluster()` instead of setting them later
3. Test connection handling - the Apache driver may have different retry/timeout behavior

---

## AWS S3

**Upgraded**: `github.com/aws/aws-sdk-go` → `github.com/aws/aws-sdk-go-v2`

### Breaking Changes

This is a **major version upgrade** with significant API changes.

#### 1. Import Paths Changed
```diff
- import "github.com/aws/aws-sdk-go/aws"
- import "github.com/aws/aws-sdk-go/aws/session"
- import "github.com/aws/aws-sdk-go/service/s3"

+ import "github.com/aws/aws-sdk-go-v2/aws"
+ import "github.com/aws/aws-sdk-go-v2/config"
+ import "github.com/aws/aws-sdk-go-v2/service/s3"
+ import "github.com/aws/aws-sdk-go-v2/service/s3/types"
```

#### 2. Client Initialization Changed
```diff
- sess := session.NewSession(&aws.Config{...})
- client := s3.New(sess)

+ cfg, err := config.LoadDefaultConfig(ctx)
+ client := s3.NewFromConfig(cfg)
```

#### 3. All Methods Now Require Context
Every S3 operation requires `context.Context` as the first parameter:

```diff
- client.PutObject(&s3.PutObjectInput{...})
+ client.PutObject(ctx, &s3.PutObjectInput{...})
```

#### 4. Type Changes
Many types have moved to the `types` package:

```diff
- ACL: aws.String("public-read")
+ ACL: types.ObjectCannedACLPublicRead

- BucketCannedACL: aws.String("public-read")
+ ACL: types.BucketCannedACLPublicRead
```

#### 5. Pointer vs Value Changes
Many fields that were pointers are now values and vice versa.

#### 6. Error Handling Changed
Error handling uses `errors.As()` instead of type assertions:

```diff
- if aerr, ok := err.(awserr.Error); ok {
-     switch aerr.Code() {
-     case s3.ErrCodeNoSuchKey:
-     }
- }

+ var nsk *types.NoSuchKey
+ if errors.As(err, &nsk) {
+     // handle
+ }
```

### Migration Guide

1. Update all import paths to v2
2. Replace `session.NewSession` with `config.LoadDefaultConfig`
3. Add context to all S3 operations
4. Update ACL and other enums to use `types` package
5. Update error handling to use `errors.As()`
6. Review pointer vs value changes
7. Test thoroughly - authentication and credentials handling changed

---

## Kafka

**Changed**: `github.com/Shopify/sarama` → `github.com/IBM/sarama`

### Breaking Changes

#### 1. Module Path Changed
The Sarama library was transferred from Shopify to IBM:

```diff
- import "github.com/Shopify/sarama"
+ import "github.com/IBM/sarama"
```

#### 2. Version Upgraded
`v1.37.2` → `v1.46.3`

### Migration Guide

1. Update all import paths from `Shopify` to `IBM`
2. No API changes - this is just a module path change
3. Review release notes between v1.37 and v1.46 for any behavior changes

---

## StatsD

**Upgraded**: `github.com/alexcesaro/statsd` → `github.com/smira/go-statsd`

### Breaking Changes

#### 1. Library Replaced
The old library (`alexcesaro/statsd`) is unmaintained since 2016. Replaced with `smira/go-statsd`.

#### 2. Import Changed
```diff
- import "github.com/alexcesaro/statsd"
+ import "github.com/smira/go-statsd"
```

#### 3. Client Initialization Changed
The initialization API is different:

```diff
- client, err := statsd.New(statsd.Address(addr))
+ client := statsd.NewClient(addr, 
+     statsd.MetricPrefix("myapp."),
+     statsd.TagStyle(statsd.TagFormatDatadog))
```

#### 4. Method Signatures May Differ
Review all StatsD method calls as signatures may have changed.

### Migration Guide

1. Update import paths
2. Update client initialization code
3. Verify all metric calls still work
4. Test metric reporting in your monitoring system

### Note

If you're using DogStatsD specifically, consider using `github.com/DataDog/datadog-go/v5/statsd` which is maintained by DataDog and supports their extended features.

---

## Echo Web Framework

**Upgraded**: `github.com/labstack/echo` v2 → `github.com/labstack/echo/v4`

### Breaking Changes

#### 1. Import Path Changed
```diff
- import "github.com/labstack/echo"
+ import "github.com/labstack/echo/v4"
```

#### 2. Engine Package Removed
Echo v4 removed the custom `engine` package and now uses standard library types directly:

```diff
- import "github.com/labstack/echo/engine"
- request engine.Request
+ // Use *http.Request directly
```

#### 3. Request/Response Access Changed
Methods became fields:

```diff
- c.Request().Method()
- c.Request().URL()
- c.Response().Status()
+ c.Request().Method  // field
+ c.Request().URL     // field
+ c.Response().Status // field
```

#### 4. Context Handling Changed
```diff
- ctx := c.StdContext()
- c.SetStdContext(ctx)
+ ctx := c.Request().Context()
+ c.SetRequest(c.Request().WithContext(ctx))
```

#### 5. URL Methods Changed
```diff
- url.QueryString()
+ url.RawQuery
```

#### 6. Host Access Changed
```diff
- request.Host()
+ request.Host
```

### Migration Guide

1. Update import to `/v4`
2. Remove any `engine` package usage
3. Change method calls to field access for Request/Response
4. Update context handling code
5. Review custom middleware - it may need updates
6. Test all routes and middleware

---

## Testing (Ginkgo/Gomega)

**Upgraded**: Ginkgo v1 → v2, Gomega v1 → v2

### Breaking Changes

#### 1. Import Paths Changed
```diff
- import . "github.com/onsi/ginkgo"
- import . "github.com/onsi/gomega"
+ import . "github.com/onsi/ginkgo/v2"
+ import . "github.com/onsi/gomega"
```

#### 2. Command-Line Flags Changed
If you run `ginkgo` CLI directly:

```diff
- ginkgo --randomizeAllSpecs
- ginkgo --randomizeSuites
+ ginkgo --randomize-all
+ ginkgo --randomize-suites
```

#### 3. Some Matchers Renamed
Check Gomega v2 release notes for any matcher renames.

#### 4. Deprecation Warnings
Ginkgo v2 has deprecation warnings. Acknowledge with:
```bash
export ACK_GINKGO_DEPRECATIONS=2.27.3
```

### Migration Guide

1. Update all test file imports to `/v2`
2. Update any Makefile or CI scripts using ginkgo CLI flags
3. Run tests and address any deprecation warnings
4. Update custom matchers if you have any

---

## Mocking (gomock)

**Upgraded**: `github.com/golang/mock` → `go.uber.org/mock`

### Breaking Changes

#### 1. Module Path Changed
Google no longer maintains `golang/mock`. Uber forked it:

```diff
- import "github.com/golang/mock/gomock"
+ import "go.uber.org/mock/gomock"
```

#### 2. mockgen Tool Changed
Install the new version:

```bash
go install go.uber.org/mock/mockgen@latest
```

### Migration Guide

1. Update all imports to `go.uber.org/mock/gomock`
2. Install new mockgen tool
3. Regenerate all mocks: `make mocks`
4. No API changes - the interfaces are compatible

---

## Other Libraries

### gRPC
**Upgraded**: `v1.46.2` → `v1.77.0`

- No breaking changes in typical usage
- Review [gRPC-Go release notes](https://github.com/grpc/grpc-go/releases) for version-specific changes
- Some internal APIs may have changed

### Viper (Configuration)
**Upgraded**: `v1.13.0` → `v1.21.0`

- No breaking changes
- New features available

### Logrus (Logging)
**Upgraded**: `v1.7.0` → `v1.9.3`

- No breaking changes
- Performance improvements

### Gorilla Mux
**Upgraded**: `v1.8.0` → `v1.8.1`

- Patch release, no breaking changes

### Eclipse Paho MQTT
**Upgraded**: `v1.1.0` → `v1.5.1`

- Review if you use MQTT
- WebSocket support added
- Some internal changes

### govalidator
**Upgraded**: `v0.0.0-20171111151018` → `v0.0.0-20230301143203`

- Latest code from 2023
- No API changes
- Note: The library has version tags (v11) but doesn't follow Go module versioning correctly

---

## Testing Your Upgrade

### Recommended Testing Strategy

1. **Unit Tests**: Run your full test suite
   ```bash
   make test-unit
   ```

2. **Integration Tests**: Test with real dependencies where applicable
   ```bash
   make test-integration
   ```

3. **Component-Specific Testing**:
   - **Redis**: Test connection pooling, pub/sub, locks
   - **PostgreSQL**: Test transactions, bulk operations
   - **MongoDB**: Test CRUD operations, aggregations
   - **Cassandra**: Test queries, batch operations
   - **S3**: Test uploads, downloads, presigned URLs
   - **Kafka**: Test producers and consumers
   - **API Endpoints**: Test all HTTP handlers if using Echo

4. **Performance Testing**: Some upgrades may affect performance
   - Monitor Redis command latency
   - Check database query performance
   - Verify S3 operation times

5. **Error Handling**: Verify error handling works correctly
   - Test context cancellation
   - Test timeout scenarios
   - Test connection failures

---

## Common Migration Patterns

### Adding Context to Existing Code

Many libraries now require context. Common pattern:

```go
// Before
func doSomething() error {
    result, err := client.Get("key")
    return err
}

// After
func doSomething(ctx context.Context) error {
    result, err := client.Get(ctx, "key")
    return err
}
```

### Handling New Return Values

Many methods now return additional values:

```go
// Before
err := db.Insert(&user)

// After
result, err := db.Insert(&user)
if err != nil {
    return err
}
// Use result if needed
log.Printf("Inserted %d rows", result.RowsAffected())
```

### Using Type-Safe Enums

Many string constants became typed enums:

```go
// Before
ACL: aws.String("public-read")

// After  
ACL: types.ObjectCannedACLPublicRead
```

---

## Getting Help

If you encounter issues:

1. Check the relevant library's migration guide:
   - [Redis v9 Migration](https://github.com/redis/go-redis/blob/master/v9-migration-guide.md)
   - [go-pg v10 Migration](https://github.com/go-pg/pg/wiki/Migration-guide)
   - [MongoDB Driver](https://www.mongodb.com/docs/drivers/go/current/)
   - [AWS SDK v2 Migration](https://aws.github.io/aws-sdk-go-v2/docs/migrating/)
   - [Echo v4 Guide](https://echo.labstack.com/docs/migrating)

2. Review the CHANGELOG in this repository

3. Search for similar issues in the upstream library repositories

4. Check this repository's GitHub issues

---

## Summary of All Breaking Changes

| Component | Old Version | New Version | Major Breaking Changes |
|-----------|-------------|-------------|------------------------|
| Redis | v8 | v9 | Context required, method changes |
| PostgreSQL | v6 | v10 | Return values, Model() required |
| MongoDB | mgo.v2 | mongo-driver/v2 | Complete API rewrite |
| Cassandra | gocql | apache/v2 | Module path, NewCluster signature |
| AWS S3 | v1 | v2 | Complete API rewrite, context required |
| Kafka/Sarama | Shopify | IBM | Module path changed |
| StatsD | alexcesaro | smira | Library replaced |
| Echo | v2 | v4 | Methods → fields, engine removed |
| Ginkgo | v1 | v2 | Import path, CLI flags |
| gomock | golang | uber-go | Module path |

---

**Last Updated**: 2025-12-10
**Extensions Version**: v10.0.0


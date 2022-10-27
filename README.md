# Go Bench Away!
### Utility to benchmark any Golang project and visualize results

As of this writing, the project is maturing from "proof of concept" to "open-source utility".
Some things may be broken, some are going to break. Also no documentations and no tests.


Before V1.0:
 * Make --speed an option for compare
 * Host info
 * Cleanup log vs printf
 * Wait on multiple jobs
 * Submit and wait
 * Single result set report
 * Documentation and examples
 * Tests
 * Less code repetition in commands
 - Spin loop in worker when server is down
 - Use template for bash script and maybe save it as artifact
 - Allow worker customization of tempdir root (where jobs directories are created)
 - Add user@host to client name
---

# Installation

Download the latest [release](https://github.com/mprimi/go-bench-away/releases) or install from source with:

```
$ go install github.com/mprimi/go-bench-away@vX.Y.Z
```

## NATS Server setup
```
$ nats-server -js -auth ${TOKEN}
```

## Schema initialization
One-time initialization (creates empty stream, key-value store, and object store):
```
$ go-bench-away -server nats://${TOKEN}@${SERVER_IP}:4222 init
2022/10/20 19:56:54 Jobs stream: default-jobs
2022/10/20 19:56:54 Jobs KV store created: default-jobs
2022/10/20 19:56:54 Artifacts Obj store: default-artifacts
```

## Worker setup
On the **benchmark host** (i.e., "bare metal" host where benchmarks will execute)
```
$ go-bench-away -server nats://${TOKEN}@${SERVER_IP}:4222 worker
2022/10/20 19:57:04 Worker ready
```
This is a long-running process, so you may want to run it inside a `screen` session, or as a daemon service

## Submit a job
Run this from anywhere: your laptop, a GitHub action, a Jenkins job, etc.
```
$ go-bench-away -server nats://${TOKEN}@${SERVER_IP}:4222 submit -remote https://github.com/nats-io/nats-server.git -ref v2.9.3 -reps 3 -tests_dir server -filter 'BenchmarkJetStreamPublish/.*/Sync'
```

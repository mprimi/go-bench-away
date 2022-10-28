# Go Bench Away!
### Utility to benchmark any Golang project and visualize results

As of this writing, the project is maturing from "proof of concept" to "open-source utility".
Some things may be broken, some are going to break. Also no documentations and no tests.


Before V1.0:
 * Run with custom Go
 * Add ranges to trend report table
 * Ensure uniqueness of refs in graph (e.g. same SHA, different GO)
 * Single result set report
 * Add time/op to compare report
 * Rename compare-speed to compare report
 * Handle re-delivery (call InProgress() ?)
 * Wait on multiple jobs
 * Add wait option to submit
 * Submit multiple
 * Documentation and examples
 * Split template to reuse style and other common elements
 * Tests
 * Less code repetition in commands
 - Spin loop in worker when server is down
 - Use template for bash script and maybe save it as artifact
 - Use template for jobs table (shared by all reports)

Future/Wishlist:
 * Embedded mode - run server and worker in-process
 * Expose `internal` as packages so parts can be used as library
 * Web UI (browse jobs)
 * Search jobs

---

# What is this?

`go-bench-away` (GBA) is a utility to orchestrate and automate execution and analysis of benchmarks. It is inspired by tools such as [Jenkins](https://www.jenkins.io) and [Genie](https://github.com/Netflix/genie).

Unlike those generic tools, GBA does just one thing: run benchmarks, produce results visual reports.

GBA works with **any Golang repository** that implements [`testing.Benchmark`](https://pkg.go.dev/testing#hdr-Benchmarks), without any change to the target repository.

It is also similar to [gobenchdata](https://github.com/bobheadxi/gobenchdata), except is designed to natively run benchmarks *elsewhere* (i.e., not on your laptop, but on a dedicated bare metal host somewhere in the cloud), hence the name go-bench-*away*.


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

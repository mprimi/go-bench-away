# Caddy

https://github.com/caddyserver/caddy.git
Benchmark.*Matcher
modules/caddyhttp
2.5.0
2.6.0
2.6.2

```sh
for ref in v2.5.0 v2.6.0 v2.6.2;
do
  go run main.go -server nats://localhost:4222 submit -ref ${ref} -remote https://github.com/caddyserver/caddy.git -reps 10 -min_runtime 3s -tests_dir modules/caddyhttp -filter 'Benchmark.*Matcher';
done
```
jobId: 5738b4d3-d872-455b-94b3-5850784a0af2
jobId: f797c22f-2e82-44bf-aae0-ec1348e13f07
jobId: 770305d0-07c7-4042-967f-a08ceff1069c

```
go run main.go -server nats://localhost:4222 trend-report  5738b4d3-d872-455b-94b3-5850784a0af2 f797c22f-2e82-44bf-aae0-ec1348e13f07 770305d0-07c7-4042-967f-a08ceff1069c
```


# Zinc

https://github.com/zinclabs/zinc.git
pkg/zutils/base62
'BenchmarkBase62.*'

0.2.0
0.2.9
0.3.3


# Gitea
https://github.com/go-gitea/gitea.git
modules/git

Benchmark.*
1.17.0
1.16.0
1.15.0
1.14.0


# go-ethereum

https://github.com/ethereum/go-ethereum.git
common/bitutil (Not much difference)
core
core/vm/runtime
master
v1.10.10

```sh
for ref in v1.10.10 master;
do
  go run main.go -server nats://localhost:4222 submit -ref ${ref} -remote https://github.com/ethereum/go-ethereum.git -reps 10 -min_runtime 1s -tests_dir core/vm/runtime -filter 'BenchmarkEVM.*';
done
```

jobId: ae654d6e-a8ab-412d-a045-1cd0411e1bf8
jobId: 1674a5ae-0801-4c98-b378-b3cc4171a335


# MinIO
https://github.com/minio/minio.git
internal/s3select
Benchmark.*(100K|1M)
RELEASE.2022-10-24T18-35-07Z
RELEASE.2022-08-02T23-59-16Z
RELEASE.2022-01-28T02-28-16Z

```sh
for ref in RELEASE.2022-01-28T02-28-16Z RELEASE.2022-08-02T23-59-16Z RELEASE.2022-10-24T18-35-07Z;
do
  go run main.go -server nats://localhost:4222 submit -ref ${ref} -remote https://github.com/minio/minio.git -reps 5 -min_runtime 1s -tests_dir internal/s3select -filter 'Benchmark.*(100K|1M)';
done
```
jobId: 7e98b78e-1445-4067-a892-e945ed046dcb
jobId: 9b5693f2-7d4a-4476-b6c0-85ac522b0672
jobId: a3ede51d-d349-4a0f-94ad-2801893f540b

go run main.go -server nats://localhost:4222 trend-report 7e98b78e-1445-4067-a892-e945ed046dcb 9b5693f2-7d4a-4476-b6c0-85ac522b0672 a3ede51d-d349-4a0f-94ad-2801893f540b

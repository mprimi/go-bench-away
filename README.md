# Go Bench Away!
### Utility to benchmark any Golang project and visualize results

As of this writing, the project is maturing from "proof of concept" to "open-source utility".
Some things may be broken, some are going to break. Also no documentations and no tests.


Before V1.0:
 * Worker tee job output with MultiWriter
 * Cleanup log vs printf
 * Namespacing
 * Wait on multiple jobs
 * Submit and wait
 * Single result set report
 * Trend (n>2) report
 * Documentation and examples
 * Tests
 * Less code repetition in commands
 - Spin loop in worker when server is down
 - Use template for bash script and maybe save it as artifact
 - Allow worker customization of tempdir root (where jobs directories are created)

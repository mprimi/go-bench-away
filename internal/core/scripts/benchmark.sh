#!/usr/bin/env bash

# This script is invoked by the worker.
#
# 1. Clone the project from a specified origin
#    - Shallow checkout (no history, small and fast)
# 2. Run go benchmarks
#    - Place the output at a conventional location

# Stop if any commands returns non-zero status
set -e
# Uncomment next line to debug this script, prints every step evaluated
# set -x

# This is invoked progragmatically, not by humans.
# It expect a long and exact series of command-line arguments.
#
# Example:
# $ ./benchmark.sh \
#   /tmp/go-bench-away/job-1234567890 \
#   /tmp/go-bench-away/job-1234567890/results.txt \
#   /tmp/go-bench-away/job-1234567890/sha.txt \
#   /tmp/go-bench-away/job-1234567890/go_version.txt \
#   https://github.com/nats-io/nats-server.git \
#   v2.9.2 \
#   server \
#   BenchmarkJetStreamPublish* \
#   5 \
#   3s \
#   180m \
#
# Path (absolute) to an (existing) temporary folder under which all work is done
ROOT_DIR="${1}"
# Path (absolute) of the file where to write benchmark results
OUTPUT_FILE="${2}"
# Path (absolute) of the file where to write the checkout commit hash
# (i.e. the SHA that GIT_REF resolves to)
SHA_FILE="${3}"
# Path (absolute) of the file where to write the go version used
GO_VERSION_FILE="${4}"
# Git remote URL to clone code from
GIT_REMOTE="${5}"
# Name of the git reference to checkout (branch, tag, SHA, ...)
GIT_REF="${6}"
# Name of the sub-directory of the projects where to run tests from
# (Use . to run tests in the source root directory)
TESTS_DIR="${7}"
# Expression to filter which benchmark tests to run (passed to `go test`)
BENCHMARKS_FILTER="${8}"
# Number of times each benchmark tests is repeated (passed to `go test`)
BENCHMARK_REPETITIONS="${9}"
# Minimum runtime of each of the benchmark tests (passed to `go test`)
BENCHMARK_MIN_RUN_TIME="${10}"
# Maximum amount of time for all benchmarks to run before timing out (passed to `go test`)
MAX_RUN_TIME="${11}"

# The following variables are hard-coded, but some of them are good candidates
# to become parameters (or allow environment-based overrides) in the future

# The go executable, must be in PATH
GO=`which go`
# The git executable, must be in PATH
GIT=`which git`
# Extra options passed to `git clone`
GIT_CLONE_OPS="--quiet --depth=1 --single-branch -c advice.detachedHead=false"
# Extra options passed to `go test`
GO_TEST_OPTS="-v"
# Name of checkout folder (within ROOT_DIR)
CHECKOUT_DIR="source.git"

###
### Helper functions
###

# Fatal error
function fail () {
  echo "❌ $*"
  exit 1
}

# Check that a given variable (passed by name) is set
function check_variable_set () {
  if [ -n "${1}" ] ; then
    local name="${1}"
    local value="${!1}"
    echo " > ${name}=${value}"
    if [ -z "${value}" ]; then
      fail "Required variable ${name} is not set"
    fi
  else
    fail "Variable check invoked with null argument"
  fi
}

# Check that a given variable (passed by name) is set and resolves to an executable
function check_executable_variable () {
  local name="${1}"
  local exec_path="${!1}"
  check_variable_set "${name}"
  if [ ! -x "${exec_path}" ]; then
    fail "Required executable not found: ${name}=${exec_path}"
  fi
}

###
### Validate arguments and environment
###

required_vars="ROOT_DIR OUTPUT_FILE SHA_FILE GO_VERSION_FILE GIT_REMOTE GIT_REF TESTS_DIR BENCHMARKS_FILTER BENCHMARK_REPETITIONS BENCHMARK_MIN_RUN_TIME MAX_RUN_TIME"
for rv in ${required_vars}; do
  check_variable_set "${rv}"
done

required_exe_vars="GO GIT"
for rv in ${required_exe_vars}; do
  check_executable_variable "${rv}"
done

test -d "${ROOT_DIR}" || fail "ROOT_DIR=${ROOT_DIR} does not exist or is not a directory"
test -e "${OUTPUT_FILE}" && fail "OUTPUT_FILE=${OUTPUT_FILE} exists"
test -e "${SHA_FILE}" && fail "SHA_FILE=${SHA_FILE} exists"
test -e "${GO_VERSION_FILE}" && fail "GO_VERSION_FILE=${GO_VERSION_FILE} exists"


mkdir -p "${ROOT_DIR}/${CHECKOUT_DIR}" || fail "Failed to create checkout directory: ${ROOT_DIR}/${CHECKOUT_DIR}"

echo

###
### Clone source
###
echo "Cloning ${GIT_REMOTE} ref: ${GIT_REF} to ${ROOT_DIR}/${CHECKOUT_DIR}"
${GIT} clone ${GIT_CLONE_OPS} --branch ${GIT_REF} ${GIT_REMOTE} "${ROOT_DIR}/${CHECKOUT_DIR}" || fail "Failed to checkout source"

cd "${ROOT_DIR}/${CHECKOUT_DIR}/${TESTS_DIR}" || fail "Failed to cd to ${ROOT_DIR}/${CHECKOUT_DIR}/${TESTS_DIR}"

echo

# Record the commit SHA
echo "SHA of ${GIT_REF}:"
${GIT} rev-parse --verify HEAD | tee "${SHA_FILE}"

test -s "${SHA_FILE}" || fail "Failed to identify commit SHA"

# Record the go version
echo "Go runtime:"
${GO} version | tee "${GO_VERSION_FILE}"

test -s "${GO_VERSION_FILE}" || fail "Failed to identify commit SHA"

echo

###
### Run benchmarks
###
echo "Running benchmarks with filter '${BENCHMARKS_FILTER}' (${BENCHMARK_REPETITIONS} repetitions, ${BENCHMARK_MIN_RUN_TIME} min runtime, timeout in ${MAX_RUN_TIME})"
${GO} test ${GO_TEST_OPTS} --bench "${BENCHMARKS_FILTER}" --run "${BENCHMARKS_FILTER}" --count ${BENCHMARK_REPETITIONS} -benchtime ${BENCHMARK_MIN_RUN_TIME} -timeout ${MAX_RUN_TIME} | tee ${OUTPUT_FILE}

test_exit_code="${PIPESTATUS[0]}"

test "${test_exit_code}" -eq 0 || fail "Non-zero exit code: ${test_exit_code}"
test -s "${OUTPUT_FILE}" || fail "Benchmarks produced no results"

echo

echo Done

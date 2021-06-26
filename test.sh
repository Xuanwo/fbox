#!/bin/sh

test_passed() {
  printf "✅ Tested PASSED\n"
  exit 0
}

test_failed() {
  printf "❌ Test FAILED\n"
  exit 1
}

ok() {
  printf "OK\n"
}

error() {
  printf "ERR\n"
  test_failed
}

random_file() {
  fn="$1"

  dd if=/dev/urandom of="$fn" bs=1024 count=$((10 * 1024)) > /dev/null 2>&1
}

shasum() {
  fn="$1"

  sha256sum "$fn" | cut -f 1 -d ' '
}

in="in"
out="out"

curl -qsSL -o - http://localhost:8000/nodes | jq '.'

printf "Generating random input file... "
random_file "$in" || error
ok

in_shasum="$(shasum "$in")"

printf "Uploading %s ..." "$in"
curl -qsSL -T "$in" "http://localhost:8000/upload/$in" || error
ok

printf "Downloading %s ..." "$in"
curl -qsSL -o "$out" "http://localhost:8000/download/$in" || error
ok

out_shasum="$(shasum "$out")"

printf "Checking integrity of %s vs %s ..." "$in" "$out"
[ "$in_shasum" -eq "$out_shasum" ] error
ok

test_passed

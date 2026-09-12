#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "== Building C library =="
gcc -shared -fPIC -O2 -o include/libcalculator.so include/c_lib/calculator.c
echo "  -> include/libcalculator.so"

echo "== Building Rust library =="
(cd include/rust_lib && cargo build --release)

# На macOS Cargo собирает .dylib, на Linux — .so
if [ -f include/rust_lib/target/release/libcalculator_rust.dylib ]; then
    cp include/rust_lib/target/release/libcalculator_rust.dylib include/
    echo "  -> include/libcalculator_rust.dylib"
else
    cp include/rust_lib/target/release/libcalculator_rust.so include/
    echo "  -> include/libcalculator_rust.so"
fi

echo "Build complete."
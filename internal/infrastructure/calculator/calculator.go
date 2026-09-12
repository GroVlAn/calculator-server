package calculator

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo linux  LDFLAGS: -L${SRCDIR}/../../../include -Wl,-rpath,'$ORIGIN' -lcalculator -lcalculator_rust
#cgo darwin LDFLAGS: -L${SRCDIR}/../../../include -Wl,-rpath,${SRCDIR}/../../../include -lcalculator -lcalculator_rust

#include "c_lib/calculator.h"
#include "calculator_rust.h"
*/
import "C"

type Calculator struct{}

func New() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b int64) int64 {
	return int64(C.add(C.int64_t(a), C.int64_t(b)))
}

func (c *Calculator) Sub(a, b int64) int64 {
	return int64(C.sub(C.int(a), C.int(b)))
}

package id

import "sync/atomic"

var current uint64

func Next() uint64 {
	return atomic.AddUint64(&current, 1)
}

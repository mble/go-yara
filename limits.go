// Copyright © 2015-2019 Hilko Bengen <bengen@hilluzination.de>
// All rights reserved.
//
// Use of this source code is governed by the license that can be
// found in the LICENSE file.

package yara

/*
#include <yara/limits.h>
*/
import "C"

// Limits contains the current limits as defined
// by yara/limits.h
type Limits struct {
	MaxPath          int
	MaxThreads       int
	MaxStringMatches int
}

// FetchLimits returns a Limits struct with the current
// limits
func FetchLimits() (*Limits, error) {
	l := &Limits{
		MaxPath:          int(C.MAX_PATH),
		MaxThreads:       int(C.YR_MAX_THREADS),
		MaxStringMatches: int(C.YR_MAX_STRING_MATCHES),
	}
	return l, nil
}

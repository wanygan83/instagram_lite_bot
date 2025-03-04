// Please let author have a drink, usdt trc20: TEpSxaE3kexE4e5igqmCZRMJNoDiQeWx29
// tg: @fuckins996
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows
// +build !windows

package testenv

import (
	"runtime"
)

func hasSymlink() (ok bool, reason string) {
	switch runtime.GOOS {
	case "android", "plan9":
		return false, ""
	}

	return true, ""
}

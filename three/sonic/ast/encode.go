// Please let author have a drink, usdt trc20: TEpSxaE3kexE4e5igqmCZRMJNoDiQeWx29
// tg: @fuckins996
/*
 * Copyright 2021 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ast

import (
	"sync"
	"unicode/utf8"
)

const (
	_MaxBuffer = 1024 // 1KB buffer size
)

func quoteString(e *[]byte, s string) {
	*e = append(*e, '"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				*e = append(*e, s[start:i]...)
			}
			*e = append(*e, '\\')
			switch b {
			case '\\', '"':
				*e = append(*e, b)
			case '\n':
				*e = append(*e, 'n')
			case '\r':
				*e = append(*e, 'r')
			case '\t':
				*e = append(*e, 't')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				*e = append(*e, `u00`...)
				*e = append(*e, hex[b>>4])
				*e = append(*e, hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		// if c == utf8.RuneError && size == 1 {
		//     if start < i {
		//         e.Write(s[start:i])
		//     }
		//     e.WriteString(`\ufffd`)
		//     i += size
		//     start = i
		//     continue
		// }
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				*e = append(*e, s[start:i]...)
			}
			*e = append(*e, `\u202`...)
			*e = append(*e, hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		*e = append(*e, s[start:]...)
	}
	*e = append(*e, '"')
}

var bytesPool = sync.Pool{}

func (self *Node) MarshalJSON() ([]byte, error) {
	buf := newBuffer()
	err := self.encode(buf)
	if err != nil {
		freeBuffer(buf)
		return nil, err
	}

	ret := make([]byte, len(*buf))
	copy(ret, *buf)
	freeBuffer(buf)
	return ret, err
}

func newBuffer() *[]byte {
	if ret := bytesPool.Get(); ret != nil {
		return ret.(*[]byte)
	} else {
		buf := make([]byte, 0, _MaxBuffer)
		return &buf
	}
}

func freeBuffer(buf *[]byte) {
	*buf = (*buf)[:0]
	bytesPool.Put(buf)
}

func (self *Node) encode(buf *[]byte) error {
	if self.IsRaw() {
		return self.encodeRaw(buf)
	}
	switch self.Type() {
	case V_NONE:
		return ErrNotExist
	case V_ERROR:
		return self.Check()
	case V_NULL:
		return self.encodeNull(buf)
	case V_TRUE:
		return self.encodeTrue(buf)
	case V_FALSE:
		return self.encodeFalse(buf)
	case V_ARRAY:
		return self.encodeArray(buf)
	case V_OBJECT:
		return self.encodeObject(buf)
	case V_STRING:
		return self.encodeString(buf)
	case V_NUMBER:
		return self.encodeNumber(buf)
	case V_ANY:
		return self.encodeInterface(buf)
	default:
		return ErrUnsupportType
	}
}

func (self *Node) encodeRaw(buf *[]byte) error {
	raw, err := self.Raw()
	if err != nil {
		return err
	}
	*buf = append(*buf, raw...)
	return nil
}

func (self *Node) encodeNull(buf *[]byte) error {
	*buf = append(*buf, bytesNull...)
	return nil
}

func (self *Node) encodeTrue(buf *[]byte) error {
	*buf = append(*buf, bytesTrue...)
	return nil
}

func (self *Node) encodeFalse(buf *[]byte) error {
	*buf = append(*buf, bytesFalse...)
	return nil
}

func (self *Node) encodeNumber(buf *[]byte) error {
	str := self.toString()
	*buf = append(*buf, str...)
	return nil
}

func (self *Node) encodeString(buf *[]byte) error {
	if self.l == 0 {
		*buf = append(*buf, '"', '"')
		return nil
	}

	quote(buf, self.toString())
	return nil
}

func (self *Node) encodeArray(buf *[]byte) error {
	if self.isLazy() {
		if err := self.skipAllIndex(); err != nil {
			return err
		}
	}

	nb := self.len()
	if nb == 0 {
		*buf = append(*buf, bytesArray...)
		return nil
	}

	*buf = append(*buf, '[')

	var s = (*linkedNodes)(self.p)
	var started bool
	if nb > 0 {
		n := s.At(0)
		if n.Exists() {
			if err := n.encode(buf); err != nil {
				return err
			}
			started = true
		}
	}

	for i := 1; i < nb; i++ {
		n := s.At(i)
		if !n.Exists() {
			continue
		}
		if started {
			*buf = append(*buf, ',')
		}
		started = true
		if err := n.encode(buf); err != nil {
			return err
		}
	}

	*buf = append(*buf, ']')
	return nil
}

func (self *Pair) encode(buf *[]byte) error {
	if len(*buf) == 0 {
		*buf = append(*buf, '"', '"', ':')
		return self.Value.encode(buf)
	}

	quote(buf, self.Key)
	*buf = append(*buf, ':')

	return self.Value.encode(buf)
}

func (self *Node) encodeObject(buf *[]byte) error {
	if self.isLazy() {
		if err := self.skipAllKey(); err != nil {
			return err
		}
	}

	nb := self.len()
	if nb == 0 {
		*buf = append(*buf, bytesObject...)
		return nil
	}

	*buf = append(*buf, '{')

	var s = (*linkedPairs)(self.p)
	var started bool
	if nb > 0 {
		n := s.At(0)
		if n.Value.Exists() {
			if err := n.encode(buf); err != nil {
				return err
			}
			started = true
		}
	}

	for i := 1; i < nb; i++ {
		n := s.At(i)
		if !n.Value.Exists() {
			continue
		}
		if started {
			*buf = append(*buf, ',')
		}
		started = true
		if err := n.encode(buf); err != nil {
			return err
		}
	}

	*buf = append(*buf, '}')
	return nil
}

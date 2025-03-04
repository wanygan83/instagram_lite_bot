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

package unit_test

import (
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

var (
	debugSyncGC  = os.Getenv("SONIC_SYNC_GC") != ""
	debugAsyncGC = os.Getenv("SONIC_NO_ASYNC_GC") == ""
)

func TestMain(m *testing.M) {
	go func() {
		if !debugAsyncGC {
			return
		}
		println("Begin GC looping...")
		for {
			runtime.GC()
			debug.FreeOSMemory()
		}
		println("stop GC looping!")
	}()
	time.Sleep(time.Millisecond)
	m.Run()
}

var jt = jsoniter.Config{
	ValidateJsonRawMessage: true,
}.Froze()

func TestCompatMarshalDefault(t *testing.T) {
	var obj = map[string]interface{}{
		"c": json.RawMessage("[\"<&>\"]"),
	}
	sout, serr := sonic.ConfigDefault.Marshal(obj)
	jout, jerr := jt.Marshal(obj)
	require.Equal(t, jerr, serr)
	require.Equal(t, string(jout), string(sout))

	// obj = map[string]interface{}{
	//     "a": json.RawMessage(" [} "),
	// }
	// sout, serr = sonic.ConfigDefault.Marshal(obj)
	// jout, jerr = json.Marshal(obj)
	// require.NotNil(t, jerr)
	// require.NotNil(t, serr)
	// require.Equal(t, string(jout), string(sout))

	obj = map[string]interface{}{
		"a": json.RawMessage("1"),
	}
	sout, serr = sonic.ConfigDefault.MarshalIndent(obj, "", "  ")
	jout, jerr = jt.MarshalIndent(obj, "", "  ")
	require.Equal(t, jerr, serr)
	require.Equal(t, string(jout), string(sout))
}

func TestCompatUnmarshalDefault(t *testing.T) {
	var sobj = map[string]interface{}{}
	var jobj = map[string]interface{}{}
	var data = []byte(`{"a":-0}`)
	var str = string(data)
	serr := sonic.ConfigDefault.UnmarshalFromString(str, &sobj)
	jerr := jt.UnmarshalFromString(str, &jobj)
	require.Equal(t, jerr, serr)
	require.Equal(t, jobj, sobj)

	x := struct{ A json.Number }{}
	y := struct{ A json.Number }{}
	data = []byte(`{"A":"1", "B":-1}`)
	serr = sonic.ConfigDefault.Unmarshal(data, &x)
	jerr = jt.Unmarshal(data, &y)
	require.Equal(t, jerr, serr)
	require.Equal(t, y, x)
}

func TestCompatEncoderDefault(t *testing.T) {
	var o = map[string]interface{}{
		"a": "<>",
		// "b": json.RawMessage(" [ ] "),
	}
	var w1 = bytes.NewBuffer(nil)
	var w2 = bytes.NewBuffer(nil)
	var enc1 = jt.NewEncoder(w1)
	var enc2 = sonic.ConfigDefault.NewEncoder(w2)

	require.Nil(t, enc1.Encode(o))
	require.Nil(t, enc2.Encode(o))
	require.Equal(t, w1.String(), w2.String())

	enc1.SetEscapeHTML(true)
	enc2.SetEscapeHTML(true)
	enc1.SetIndent("", "  ")
	enc2.SetIndent("", "  ")
	require.Nil(t, enc1.Encode(o))
	require.Nil(t, enc2.Encode(o))
	require.Equal(t, w1.String(), w2.String())

	enc1.SetEscapeHTML(false)
	enc2.SetEscapeHTML(false)
	enc1.SetIndent("", "")
	enc2.SetIndent("", "")
	require.Nil(t, enc1.Encode(o))
	require.Nil(t, enc2.Encode(o))
	require.Equal(t, w1.String(), w2.String())
}

func TestCompatDecoderDefault(t *testing.T) {
	var o1 = map[string]interface{}{}
	var o2 = map[string]interface{}{}
	var s = `{"a":"b"} {"1":"2"} a {}`
	var w1 = bytes.NewBuffer([]byte(s))
	var w2 = bytes.NewBuffer([]byte(s))
	var enc1 = jt.NewDecoder(w1)
	var enc2 = sonic.ConfigDefault.NewDecoder(w2)

	require.Equal(t, enc1.More(), enc2.More())
	require.Nil(t, enc1.Decode(&o1))
	require.Nil(t, enc2.Decode(&o2))
	require.Equal(t, w1.String(), w2.String())

	require.Equal(t, enc1.More(), enc2.More())
	require.Nil(t, enc1.Decode(&o1))
	require.Nil(t, enc2.Decode(&o2))
	require.Equal(t, w1.String(), w2.String())

	require.Equal(t, enc1.More(), enc2.More())
	require.NotNil(t, enc1.Decode(&o1))
	require.NotNil(t, enc2.Decode(&o2))
	require.Equal(t, w1.String(), w2.String())

	// require.Equal(t, enc1.More(), enc2.More())
	// require.NotNil(t, enc1.Decode(&o1))
	// require.NotNil(t, enc2.Decode(&o2))
	// require.Equal(t, w1.String(), w2.String())
}

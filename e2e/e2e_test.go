package e2e

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
)

// FuzzE2E creates instances of the Basic protobuf from e2e.proto, and ensures that marshalling and
// unmarshalling each instance using both protojson and the generated MarshalJson method, results in
// the same protobuf.
func FuzzE2E(f *testing.F) {
	// Seed items
	f.Add("fuzz", int32(1234), true, []byte("fuzz"), int64(1234))
	f.Add("hello world", int32(0), false, []byte("asdf"), int64(0))
	f.Add("", int32(2^32), false, []byte(""), int64(2^64))
	f.Fuzz(func(t *testing.T, a string, b int32, c bool, d []byte, e int64) {
		basic := &Basic{
			A: a,
			B: b,
			C: c,
			D: d,
			E: e,
		}

		basicMarshalUsingJson, err := json.Marshal(basic)
		if err != nil {
			t.Errorf("unable to marshal (json): %#v, %v", basic, err)
			t.FailNow()
		}
		basicUnmarshalledUsingJson := &Basic{}
		err = json.Unmarshal(basicMarshalUsingJson, basicUnmarshalledUsingJson)
		if err != nil {
			t.Errorf("unable to unmarshal (json): %q, %v", basicMarshalUsingJson, err)
			t.FailNow()
		}

		// TODO: protojson doesn't handle strings that have invalid utf-8 characters in them, but
		// encoding/json does.
		basicMarshalUsingProtoJson, err := protojson.Marshal(basic)
		if err != nil {
			t.Errorf("unable to marshal (protojson): %#v, %v", basic, err)
			t.FailNow()
		}
		basicUnmarshalledUsingProtoJson := &Basic{}
		err = protojson.Unmarshal(basicMarshalUsingProtoJson, basicUnmarshalledUsingProtoJson)
		if err != nil {
			t.Errorf("unable to unmarshal (protojson): %q, %v", basicUnmarshalledUsingProtoJson, err)
			t.FailNow()
		}

		if basicUnmarshalledUsingJson.String() != basicUnmarshalledUsingProtoJson.String() {
			t.Errorf("no match between json and protojson: %s, %s", basicUnmarshalledUsingJson.String(), basicUnmarshalledUsingProtoJson.String())
			t.FailNow()
		}
	})
}

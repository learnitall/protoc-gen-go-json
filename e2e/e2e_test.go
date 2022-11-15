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
	f.Add("fuzz", []byte("fuzz"), true)
	f.Add("hello world", []byte("asdf"), false)
	f.Add("", []byte(""), true)
	f.Add("", []byte{}, false)
	f.Fuzz(func(t *testing.T, a string, b []byte, c bool) {
		basic := &Basic{
			A: a,
			B: b,
			C: c,
		}

		basicMarshalUsingProtoJson, err := protojson.Marshal(basic)
		if err != nil {
			t.Logf("unable to marshal (protojson): %#v, %v", basic, err)
			t.SkipNow()
		}
		basicMarshalUsingJson, err := json.Marshal(basic)
		if err != nil {
			t.Errorf("unable to marshal (json): %#v, %v", basic, err)
			t.FailNow()
		}

		basicUnmarshalledUsingProtoJson := &Basic{}
		err = protojson.Unmarshal(basicMarshalUsingProtoJson, basicUnmarshalledUsingProtoJson)
		if err != nil {
			t.Logf("unable to unmarshal (protojson): %q, %v", basicUnmarshalledUsingProtoJson, err)
			t.SkipNow()
		}
		basicUnmarshalledUsingJson := &Basic{}
		err = json.Unmarshal(basicMarshalUsingJson, basicUnmarshalledUsingJson)
		if err != nil {
			t.Errorf("unable to unmarshal (json): %q, %v", basicMarshalUsingJson, err)
			t.FailNow()
		}

		if basicUnmarshalledUsingJson.String() != basicUnmarshalledUsingProtoJson.String() {
			t.Errorf("no match between json and protojson: %s, %s", basicUnmarshalledUsingJson.String(), basicUnmarshalledUsingProtoJson.String())
			t.FailNow()
		}
	})
}

package json

import (
	"encoding/json"
	"github.com/mailru/easyjson"
	"testing"
)

var str = []byte(`  { 
	"someSillyObj": { "nice": "waste of time" },  "Nested": {   "Amazing"
		: "yeah i know" },
	"bad": "missing", "Food":    
	 "i dont believe it wow", 
	 "Name":"world" ,
	 "Tags"   :  {   "a": "lol", "b": "yay" } }  `)

func TestUnmarshal(t *testing.T) {
	obj := &TestType{Tags: map[string]string{"temp": "temp"}}
	err := Unmarshal(str, obj)

	t.Logf("obj: %#v, err: %s", obj, err)
	if obj.Name != "world" {
		t.Error("name is not right")
	}
}

func TestJsonUnmarshal(t *testing.T) {
	obj := &TestType{Tags: map[string]string{"temp": "temp"}}
	err := json.Unmarshal(str, obj)

	t.Logf("obj: %#v, err: %s", obj, err)
	if obj.Name != "world" {
		t.Error("name is not right")
	}
}

func TestEasyUnmarshal(t *testing.T) {
	obj := &EasyType{Tags: map[string]string{"temp": "temp"}}
	err := easyjson.Unmarshal(str, obj)

	t.Logf("obj: %#v, err: %s", obj, err)
	if obj.Name != "world" {
		t.Error("name is not right")
	}
}

func BenchmarkNewParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			t := &TestType{}
			for pb.Next() {
				Unmarshal(str, t)
			}
		},
	)
}

func BenchmarkOldParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			t := &TestType{}
			for pb.Next() {
				json.Unmarshal(str, t)
			}
		},
	)
}

func BenchmarkEasyParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			t := &EasyType{}
			for pb.Next() {
				easyjson.Unmarshal(str, t)
			}
		},
	)
}

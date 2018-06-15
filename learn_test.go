package json

import (
	"encoding/json"
	"github.com/mailru/easyjson"
	"reflect"
	"testing"
)

var str = []byte(`  { 
	"someSillyObj": { "nice": "waste of time" },  "Nested": {   "Amazing"
		: "yeah i know" },
	"bad": "missing", "Food":    
	 "i dont believe it wow",   "EmptyList": [],
	 "Tags"   :  {   "a": "lol", "b": "yay" } ,

	 "SomeList": [ "yay", "huge suuuuuccess", "its big", "wow", "im amazed"],
	 "Name":"world"  } `)

var strNoMap = []byte(`  { 
	"someSillyObj": { "nice": "waste of time" },  "Nested": {   "Amazing"
		: "yeah i know" },
	"bad": "missing", "Food":    
	 "i dont believe it wow",    "EmptyList": [],
	 "SomeList": [ "yay", "huge suuuuuccess", "its big", "wow", "im amazed"],
	 "Name":"world" 

	 }  `)

func TestQuickScan(t *testing.T) {
	ids := quickScan(str)
	for _, id := range ids {
		if id[2] == 1 {
			t.Log(id, "fancy", string(str[id[0]:id[1]]))
		} else {
			t.Log(id, "raw", string(str[id[0]:id[1]]))
		}
	}
}

func TestUnmarshal(t *testing.T) {
	obj := &TestType{SomeList: []string{"im already here"}, Tags: map[string]string{"temp": "temp"}}
	err := Unmarshal(str, obj)

	obj2 := &TestType{SomeList: []string{"im already here"}, Tags: map[string]string{"temp": "temp"}}
	err2 := json.Unmarshal(str, obj2)

	t.Logf("obj: \n%#v, err: %s", obj, err)
	t.Logf("obj2: \n%#v, err2: %s", obj2, err2)

	t.Logf("obj.Nested: \n%#v, err: %s", obj.Nested, err)
	t.Logf("obj2.Nested: \n%#v, err2: %s", obj2.Nested, err2)
	if !reflect.DeepEqual(obj, obj2) || !reflect.DeepEqual(err, err2) {
		t.Error("different outcomes")

	}
}

func TestEasyUnmarshal(t *testing.T) {
	obj := &EasyType{SomeList: []string{"im already here"}, Tags: map[string]string{"temp": "temp"}}
	err := easyjson.Unmarshal(str, obj)

	t.Logf("obj: %#v, err: %s", obj, err)
	if obj.Name != "world" {
		t.Error("name is not right")
	}
}

func BenchmarkNewSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &TestType{SomeList: []string{"already in"}}
		Unmarshal(strNoMap, t)
	}
}

func BenchmarkOldSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &TestType{SomeList: []string{"already in"}}
		json.Unmarshal(strNoMap, t)
	}
}

func BenchmarkEasySingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &EasyType{SomeList: []string{"already in"}}
		easyjson.Unmarshal(strNoMap, t)
	}
}

func BenchmarkNewParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				t := &TestType{}
				Unmarshal(str, t)
			}
		},
	)
}

func BenchmarkOldParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				t := &TestType{}
				json.Unmarshal(str, t)
			}
		},
	)
}

func BenchmarkEasyParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				t := &EasyType{}
				easyjson.Unmarshal(str, t)
			}
		},
	)
}

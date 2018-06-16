package json

import (
	"encoding/json"
	"github.com/json-iterator/go"
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

	 	"SurpriseMe":      [  {}, "", "wow"   ],

	 "SomeList": [ "yay", "huge suuuuuccess", "its big", "wow", "im amazed"],
	 "Name":"world"  } `)

var str1 = []byte(`  { 
	"bad": "missing", "Food":    
	 "i dont believe it wow",   "EmptyList": [],
	 "Tags"   :  {   "a": "lol", "b": "yay" } ,
	 "SomeList": [ "yay", "huge suuuuuccess", "its big", "wow", "im amazed"
	 , " but with loads", "of", "strings", "in", "the", "list", "a", "b", "c", 
	 "d", "e", "f", "g", "h" ] } `)

var str2 = []byte(`  { 
	"someSillyObj": {},   "EmptyList": [],
	 "Tags"   :  {   "a": "lol", "b": "yay", "c": "d", "e": "f", "g": "h", "i": "j" } ,
	 "Name":"world"  } `)

var str3 = []byte(`  { 
	 "SurpriseMe":      [  {}, "", "wow", "lots of",
	 	 ["crazy stuff in", {"suddenly": "an object"}], {"here":"today"}  ],
	 "Name":"world"  } `)

var strOnlyMap = []byte(`
{  "Tags"   :  {   "a": "lol", "b": "yay" }
}

`)

var strWithList = []byte(`  { 
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

func TestInterface(t *testing.T) {
	interfaceStr := []byte(`["hello"]`)
	var oldJson, newJson interface{}
	json.Unmarshal(interfaceStr, &oldJson)
	Unmarshal(interfaceStr, &newJson)

	t.Logf("old: \n%#v", oldJson)
	t.Logf("new: \n%#v", newJson)
	if !reflect.DeepEqual(oldJson, newJson) {
		t.Error("different outcomes")

	}
}

func TestCorrectnessMixed(t *testing.T) {
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

type BrokenObject struct {
	// SomeTime time.Time
	// SomeNumber int
	// SomeChannel chan int
	// SomeFunc func()
}

func TestReporting(t *testing.T) {
	var obj *TestType
	d := Standard.ReportPlan(&obj)
	t.Log(d.String())

	{
		var obj ***[][]string
		d := Standard.ReportPlan(&obj)
		t.Log(d.String())
	}

	{
		var obj string
		d := Standard.ReportPlan(&obj)
		t.Log(d.String())
	}

}

func TestCorrectness3(t *testing.T) {
	obj := &TestType{SomeList: []string{"im already here"}, Tags: map[string]string{"temp": "temp"}}
	err := Unmarshal(str3, obj)

	obj2 := &TestType{SomeList: []string{"im already here"}, Tags: map[string]string{"temp": "temp"}}
	err2 := json.Unmarshal(str3, obj2)

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

func BenchmarkQuickScan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		quickScan(str)
	}
}

func BenchmarkNewSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &TestType{SomeList: []string{"already in"}}
		Unmarshal(strWithList, t)
	}
}

func BenchmarkOldSingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &TestType{SomeList: []string{"already in"}}
		json.Unmarshal(strWithList, t)
	}
}

func BenchmarkEasySingle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &EasyType{SomeList: []string{"already in"}}
		easyjson.Unmarshal(strWithList, t)
	}
}

func BenchmarkNewParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			var t *TestType
			for pb.Next() {
				t = &TestType{}
				Unmarshal(str, t)
				t = &TestType{}
				Unmarshal(str1, t)
				t = &TestType{}
				Unmarshal(str2, t)
				t = &TestType{}
				Unmarshal(str3, t)
			}
		},
	)
}

func BenchmarkOldParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			var t *TestType
			for pb.Next() {
				t = &TestType{}
				json.Unmarshal(str, t)
				t = &TestType{}
				json.Unmarshal(str1, t)
				t = &TestType{}
				json.Unmarshal(str2, t)
				t = &TestType{}
				json.Unmarshal(str3, t)
			}
		},
	)
}

var jsoni = jsoniter.ConfigFastest

func BenchmarkIterParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			var t *TestType
			for pb.Next() {
				t = &TestType{}
				jsoni.Unmarshal(str, t)
				t = &TestType{}
				jsoni.Unmarshal(str1, t)
				t = &TestType{}
				jsoni.Unmarshal(str2, t)
				t = &TestType{}
				jsoni.Unmarshal(str3, t)
			}
		},
	)
}

func BenchmarkEasyParallel(b *testing.B) {
	b.RunParallel(
		func(pb *testing.PB) {
			var t *EasyType
			for pb.Next() {
				t = &EasyType{}
				easyjson.Unmarshal(str, t)
				t = &EasyType{}
				easyjson.Unmarshal(str1, t)
				t = &EasyType{}
				easyjson.Unmarshal(str2, t)
				t = &EasyType{}
				easyjson.Unmarshal(str3, t)
			}
		},
	)
}

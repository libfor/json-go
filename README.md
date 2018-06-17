# json-go
Uses reflection to read and write JSON with nearly the performance of codegen

# usage

```
var myDest interface{}
var myData []byte = []byte(`  { "foo"  : ["bar",  "baz"] , "nested":{  "of":"course"}}   `)
json.Unmarshal(myData, &myDest)
```

# report plan

allows you to see the decoding plan for any given type, similar to the sql concept of "EXPLAIN"

```
Here's how I plan to decode **json.testType
Check to see if I have a nil **json.testType
  If so, create a *json.testType
Check to see if I have a nil *json.testType
  If so, create a json.testType
Look for a {, then repeatedly:
  Get a key by scanning for raw bytes
  Binary search for that key through 21 handlers
  If the key is like "EmptyList", I'll:
    Search for [, returning if I find } or ]
    Repeatedly...
      Search for ", returning if I find } or ]
      Search for closing "
      Create a string in the base from the bytes I found
    Write that new array into the pointer
  If the key is like "Food", I'll:
    Search for ", returning if I find } or ]
    Search for closing "
    Create a string in the base from the bytes I found
  If the key is like "Name", I'll:
    Search for ", returning if I find } or ]
    Search for closing "
    Create a string in the base from the bytes I found
  If the key is like "Nested", I'll:
    Check to see if I have a nil *json.nested
      If so, create a json.nested
    Look for a {, then repeatedly:
      Get a key by scanning for raw bytes
      Binary search for that key through 3 handlers
      If the key is like "Amazing", I'll:
        Search for ", returning if I find } or ]
        Search for closing "
        Create a string in the base from the bytes I found
      If it's any other key, I'll:
        If I get a {, I'll pass it off as a map[string]interface{}
        If I get a [, I'll pass it off as a []interface{}
        If I get a ", I'll pass it off as a string
        I'll dereference the result into the interface{} in the base pointer
  If the key is like "SomeList", I'll:
    Search for [, returning if I find } or ]
    Repeatedly...
      Search for ", returning if I find } or ]
      Search for closing "
      Create a string in the base from the bytes I found
    Write that new array into the pointer
  If the key is like "SurpriseMe", I'll:
    If I get a {, I'll pass it off as a map[string]interface{}
    If I get a [, I'll pass it off as a []interface{}
    If I get a ", I'll pass it off as a string
    I'll dereference the result into the interface{} in the base pointer
  If the key is like "Tags", I'll:
    Look for a {, create a map[string]string, then repeatedly:
      To get a key, I'll:
        Search for ", returning if I find } or ]
        Search for closing "
        Create a string in the base from the bytes I found
      To get a value, I'll:
        Search for ", returning if I find } or ]
        Search for closing "
        Create a string in the base from the bytes I found
      Save it in the map
    Store the new map in the base pointer
  If it's any other key, I'll:
    If I get a {, I'll pass it off as a map[string]interface{}
    If I get a [, I'll pass it off as a []interface{}
    If I get a ", I'll pass it off as a string
    I'll dereference the result into the interface{} in the base pointer
```

# unmarshalling benchmarks

Run using: `go test -bench=. -run="none" -benchmem -cpu=2`, note that EasyJson has a code-generation step and does not use reflection

Serial benchmark

```
BenchmarkSerially_Libfor-2       	 1000000	      1921 ns/op	     496 B/op	      14 allocs/op
BenchmarkSerially_StdlibJson-2   	  200000	      6644 ns/op	     912 B/op	      26 allocs/op
BenchmarkSerially_Iterjson-2     	 1000000	      2145 ns/op	     704 B/op	      23 allocs/op
BenchmarkSerially_Easyjson-2     	 1000000	      1817 ns/op	     592 B/op	      15 allocs/op
BenchmarkSerially_Jzon-2         	  300000	      5009 ns/op	    2824 B/op	      64 allocs/op
```

Parallel benchmark

```
BenchmarkParallel_Libfor-2       	  200000	      7513 ns/op	    5488 B/op	     124 allocs/op
BenchmarkParallel_StdlibJson-2   	  100000	     18807 ns/op	    6672 B/op	     168 allocs/op
BenchmarkParallel_Iterjson-2     	  200000	      7589 ns/op	    5608 B/op	     161 allocs/op
BenchmarkParallel_Easyjson-2     	  300000	      5229 ns/op	    4832 B/op	     103 allocs/op
BenchmarkParallel_Jzon-2         	  100000	     17108 ns/op	   12603 B/op	     275 allocs/op
```

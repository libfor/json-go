# json-go
Uses reflection to read and write JSON with nearly the performance of codegen

# unmarshalling benchmarks

Note, "Easyjson" uses code-gen

Run using: go test -bench=. -run="none" -benchmem -cpu=1,3

Serial benchmark

```
BenchmarkSerially_Libfor         	 1000000	      1945 ns/op	     496 B/op	      14 allocs/op
BenchmarkSerially_Libfor-3       	 1000000	      1875 ns/op	     496 B/op	      14 allocs/op
BenchmarkSerially_StdlibJson     	  200000	      6392 ns/op	     912 B/op	      26 allocs/op
BenchmarkSerially_StdlibJson-3   	  200000	      6192 ns/op	     912 B/op	      26 allocs/op
BenchmarkSerially_Iterjson       	 1000000	      2154 ns/op	     704 B/op	      23 allocs/op
BenchmarkSerially_Iterjson-3     	 1000000	      2031 ns/op	     704 B/op	      23 allocs/op
BenchmarkSerially_Easyjson       	 1000000	      1628 ns/op	     592 B/op	      15 allocs/op
BenchmarkSerially_Easyjson-3     	 1000000	      1534 ns/op	     592 B/op	      15 allocs/op
```

Parallel benchmark

```
BenchmarkParallel_Libfor         	  100000	     12445 ns/op	    5488 B/op	     124 allocs/op
BenchmarkParallel_Libfor-3       	  300000	      4722 ns/op	    5488 B/op	     124 allocs/op
BenchmarkParallel_StdlibJson     	   50000	     31767 ns/op	    6672 B/op	     168 allocs/op
BenchmarkParallel_StdlibJson-3   	  200000	     11707 ns/op	    6672 B/op	     168 allocs/op
BenchmarkParallel_Iterjson       	  100000	     13975 ns/op	    5608 B/op	     161 allocs/op
BenchmarkParallel_Iterjson-3     	  300000	      5256 ns/op	    5608 B/op	     161 allocs/op
BenchmarkParallel_Easyjson       	  200000	      8999 ns/op	    4832 B/op	     103 allocs/op
BenchmarkParallel_Easyjson-3     	  500000	      3497 ns/op	    4832 B/op	     103 allocs/op
```

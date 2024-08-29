# go-orderbook-v2

An efficient implementation of an order book in Go, using a combination of heaps and linked lists.

## Benchmarks

```
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkBidBookInsertWorstCase-8   	 1000000	      1415 ns/op	     362 B/op	       4 allocs/op
BenchmarkBidBookInsertAverage-8     	 1975744	       664.0 ns/op	     159 B/op	       2 allocs/op
PASS
ok  	orderbook	3.856s
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

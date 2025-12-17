# gcswiss

[![GoDoc](https://godoc.org/github.com/philpearl/gcswiss?status.svg)](https://godoc.org/github.com/philpearl/gcswiss) 


gcswiss is an off-heap hashmap following the swiss-tables design. Keys must
always be strings, but the values are generic. Keys & values are copied and kept
off heap, so if the values contain pointers to heap objects they must be kept
alive separately.
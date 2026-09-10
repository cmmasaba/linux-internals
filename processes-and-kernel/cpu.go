package main

func fibonacci(n uint) uint64 {
	if n < 2 {
		return uint64(n)
	}

	return fibonacci(n-2) + fibonacci(n-1)
}

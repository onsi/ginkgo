package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	f, _ := os.Create("file-output")
	for i := 0; i < 300; i++ {
		fmt.Fprintf(os.Stdout, "STDOUT %d\n", i)
		fmt.Fprintf(os.Stderr, "STDERR %d\n", i)
		fmt.Fprintf(f, "FILE %d\n", i)
		time.Sleep(10 * time.Millisecond)
	}
	f.Close()
}

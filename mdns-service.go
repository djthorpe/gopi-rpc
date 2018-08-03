package main

import (
	"fmt"
	"os"
	"time"
)

func Main() int {
	if listener, err := NewListener(nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return -1
	} else {
		fmt.Println("START")
		time.Sleep(100 * time.Second)
		fmt.Println("END")
		listener.Close()
		return 0
	}
}

func main() {
	os.Exit(Main())
}

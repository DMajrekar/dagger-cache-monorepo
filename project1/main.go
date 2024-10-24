// Simple go main that prints starting, sleeps for 10 seconds, then prints done and exits
package main

import (
	"fmt"
	"time"

	"github.com/dmajrekar/dagger-cache-monorepo/lib"
)

func main() {
	fmt.Println("Starting...")

	lib.Sleep(10 * time.Second)

	fmt.Println("Done.")
}

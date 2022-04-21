package main

import (
	"fmt"
	"github.com/paulgoleary/bor-suite/partition"
	"os"
	"strconv"
)

func main() {
	argsOnly := os.Args[1:]

	// TODO: Cobra or some such thing ...?

	var err error
	switch argsOnly[0] {
	case "create":
		{
			if err = partition.POCPartitionDatabase(argsOnly[0], 2,
				func(r string) {
					println(r)
				}); err != nil {
				panic(err)
			}
		}
	case "check":
		{
			checks := 10_000
			if len(argsOnly) > 2 {
				if checks, err = strconv.Atoi(argsOnly[2]); err != nil {
					panic(err)
				}
			}
			if err = partition.CheckPOCPartitionedDatabase(argsOnly[0], argsOnly[1], checks,
				func(r string) {
					println(r)
				}); err != nil {
				panic(err)
			}

		}
	default:
		panic(fmt.Errorf("unknown command: %v", argsOnly[0]))
	}
}

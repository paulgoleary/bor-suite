package main

import (
	"fmt"
	"github.com/paulgoleary/bor-suite/partition"
	"os"
	"strconv"
	"strings"
)

func main() {
	argsOnly := os.Args[1:]

	// TODO: Cobra or some such thing ...?

	var err error
	switch argsOnly[0] {
	case "create":
		{
			if err = partition.POCPartitionDatabase(argsOnly[1], 2,
				func(r string) {
					println(r)
				}); err != nil {
				panic(err)
			}
		}
	case "check":
		{
			checks := 10_000
			if len(argsOnly) > 4 {
				if checks, err = strconv.Atoi(argsOnly[3]); err != nil {
					panic(err)
				}
			}
			// args are: source_path check_path.0:check_path.1 freezer_path
			checkDbPaths := strings.Split(argsOnly[2], ":")
			if err = partition.CheckPOCPartitionedDatabase(argsOnly[1], argsOnly[3], checkDbPaths, checks,
				func(r string) {
					println(r)
				}); err != nil {
				panic(err)
			}
		}
	case "validate":
		{
			checkDbPaths := strings.Split(argsOnly[1], ":")
			if err = partition.ValidatePOCPartitionedDatabase(checkDbPaths, argsOnly[2],
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

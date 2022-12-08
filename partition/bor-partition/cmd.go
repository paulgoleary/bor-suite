package main

import (
	"flag"
	"fmt"
	"github.com/paulgoleary/bor-suite/partition"
	"os"
	"strings"
)

func main() {

	create := flag.NewFlagSet("create", flag.ExitOnError)
	check := flag.NewFlagSet("check", flag.ExitOnError)
	validate := flag.NewFlagSet("validate", flag.ExitOnError)

	command := getCommandOrExit(create, check, validate)

	var err error
	switch command {
	case create.Name():
		var sourcePath, targetPath string
		if sourcePath, targetPath, err = processCreateArgs(create); err != nil {
			usageAndExit(create, err)
		}
		if err = partition.POCPartitionDatabase2(sourcePath, targetPath, 2,
			func(r string) {
				println(r)
			}); err != nil {
			panic(err)
		}

	case check.Name():
		var checks int
		var sourcePath, freezerPath string
		var checkPaths []string
		if checks, sourcePath, checkPaths, freezerPath, err = processCheckArgs(check); err != nil {
			usageAndExit(check, err)
		}
		if err = partition.CheckPOCPartitionedDatabase(sourcePath, freezerPath, checkPaths, checks,
			func(r string) {
				println(r)
			}); err != nil {
			panic(err)
		}

	case validate.Name():
		var sourcePath string
		var checkPaths []string
		if sourcePath, checkPaths, err = processValidateArgs(validate); err != nil {
			usageAndExit(validate, err)
		}
		if err = partition.ValidatePOCPartitionedDatabase(checkPaths, sourcePath,
			func(r string) {
				println(r)
			}); err != nil {
			panic(err)
		}

	default:
		panic(fmt.Errorf("unknown command: %v", command))
	}
}

func processCreateArgs(flags *flag.FlagSet) (sourcePath, targetPath string, err error) {
	flags.StringVar(&sourcePath, "source", "", "source path")
	flags.StringVar(&targetPath, "target", "", "target path")

	flags.Parse(os.Args[2:])

	if _, err = os.Stat(sourcePath); err != nil {
		return
	}
	if _, err = os.Stat(targetPath); err != nil {
		return
	}
	return
}

func processCheckArgs(flags *flag.FlagSet) (checks int, sourcePath string, checkPaths []string, freezerPath string, err error) {
	flags.IntVar(&checks, "num-checks", 10_000, "number of checks to perform")
	var checkPathsStr string
	flags.StringVar(&checkPathsStr, "check", "", "check paths, separated by ':'")
	flags.StringVar(&freezerPath, "freezer", "", "freezer path")
	flags.StringVar(&sourcePath, "source", "", "source path")

	flags.Parse(os.Args[2:])

	if _, err = os.Stat(sourcePath); err != nil {
		return
	}
	if _, err = os.Stat(freezerPath); err != nil {
		return
	}
	checkPaths = strings.Split(checkPathsStr, ":")
	for _, checkPath := range checkPaths {
		if _, err = os.Stat(checkPath); err != nil {
			return
		}
	}
	return
}

func processValidateArgs(flags *flag.FlagSet) (sourcePath string, checkPaths []string, err error) {
	var checkPathsStr string
	flags.StringVar(&checkPathsStr, "check", "", "check paths, separated by ':'")
	flags.StringVar(&sourcePath, "source", "", "source path")

	flags.Parse(os.Args[2:])

	if _, err = os.Stat(sourcePath); err != nil {
		return
	}
	checkPaths = strings.Split(checkPathsStr, ":")
	for _, checkPath := range checkPaths {
		if _, err = os.Stat(checkPath); err != nil {
			return
		}
	}
	return
}

func getCommandOrExit(commands ...*flag.FlagSet) string {
	if len(os.Args) < 2 {
		names := make([]string, len(commands))
		for i, command := range commands {
			names[i] = command.Name()
		}
		fmt.Printf("No command specified, expecting one of: %v\n", strings.Join(names, ", "))
		os.Exit(1)
	}
	return os.Args[1]
}

func usageAndExit(set *flag.FlagSet, err error) {
	fmt.Printf("%v command ERROR: %v\n", set.Name(), err)
	set.Usage()
	os.Exit(1)
}

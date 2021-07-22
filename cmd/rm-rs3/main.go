package main

import (
	"flag"
	"fmt"
	"os"

	s3mpty "github.com/dacort/s3mpty/internal/s3mpty"
)

const (
	defaultDryRun = false
)

var dryRun bool

func init() {
	const (
		defaultDryRun = false
	)
	flag.BoolVar(&dryRun, "dryrun", defaultDryRun, "Display the operations that would be performed without actually running them.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-dryrun] <bucket_name>\n", os.Args[0])

		flag.PrintDefaults()
	}
}

func checkArgs() string {
	bucket_name := flag.Arg(0)
	if bucket_name == "" {
		fmt.Println("Error: must provide bucket name as first argument.")
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() > 1 {
		fmt.Println("Error: Only one command-line argument allowed, found: ", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

	return bucket_name
}

func main() {
	flag.Parse()
	bucket_name := checkArgs()

	sess := s3mpty.NewSession()
	client := s3mpty.NewClient(sess, bucket_name)

	// deleted_objects := s3mpty.DeleteObjectsFromBucket(client, bucket_name, dryRun)
	deleted_versions := s3mpty.DeleteVersionsFromBucket(client, bucket_name, dryRun)

	if dryRun {
		fmt.Println("(dryrun) Deleted", deleted_versions, "versions.")
	} else {
		fmt.Println("Deleted", deleted_versions, "versions.")
	}
}

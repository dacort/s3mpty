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

var (
	dryRun bool
	prefix string
)

func init() {
	flag.BoolVar(&dryRun, "dryrun", defaultDryRun, "Display the operations that would be performed without actually running them.")
	flag.StringVar(&prefix, "prefix", "", "Only delete objects with the specified prefix.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-dryrun] [-prefix <prefix>] <bucket_name>\n", os.Args[0])

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

	// deleted_objects := s3mpty.DeleteObjectsFromBucket(client, bucket_name, prefix, dryRun)
	deleted_versions := s3mpty.DeleteVersionsFromBucket(client, bucket_name, prefix, dryRun)

	// Construct the output message
	dryRunPrefix := ""
	if dryRun {
		dryRunPrefix = "(dryrun) "
	}
	
	if prefix != "" {
		fmt.Printf("%sDeleted %d versions with prefix '%s'.\n", dryRunPrefix, deleted_versions, prefix)
	} else {
		fmt.Printf("%sDeleted %d versions.\n", dryRunPrefix, deleted_versions)
	}
}

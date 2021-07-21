package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

func getBucketRegion(svc *s3.S3, bucket_name string) string {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket_name),
	}

	result, err := svc.GetBucketLocation(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Fatal(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatal(err.Error())
		}
	}
	return *result.LocationConstraint
}

func main() {
	flag.Parse()
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

	// We use SharedConfigState so we can make use of credential_process
	// Note: This is potentially unsafe
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		// handle error
		log.Fatal("Could not load credentials", err)
	}

	region_name := aws.String(getBucketRegion(s3.New(sess), bucket_name))
	svc := s3.New(sess, aws.NewConfig().WithRegion(*region_name))
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket_name),
	}

	counter := 0
	err = svc.ListObjectsV2Pages(input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			counter += int(*page.KeyCount)

			delete_input := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket_name),
				Delete: &s3.Delete{Objects: []*s3.ObjectIdentifier{}},
			}
			for _, obj := range page.Contents {
				if dryRun {
					fmt.Printf("(dryrun) delete: s3://%s/%s\n", bucket_name, *obj.Key)
				} else {
					delete_input.Delete.Objects = append(delete_input.Delete.Objects, &s3.ObjectIdentifier{Key: obj.Key})
				}

			}
			if !dryRun {
				svc.DeleteObjects(delete_input)
			}

			return lastPage
		})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	version_counter := 0
	version_input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket_name),
	}
	svc.ListObjectVersionsPages(version_input,
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {

			delete_input := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket_name),
				Delete: &s3.Delete{Objects: []*s3.ObjectIdentifier{}},
			}
			version_counter += len(page.DeleteMarkers)
			for _, obj := range page.DeleteMarkers {
				if dryRun {
					fmt.Printf("(dryrun) delete version: s3://%s/%s#%s\n", bucket_name, *obj.Key, *obj.VersionId)
				} else {
					delete_input.Delete.Objects = append(delete_input.Delete.Objects, &s3.ObjectIdentifier{Key: obj.Key, VersionId: obj.VersionId})
				}
			}
			if !dryRun {
				svc.DeleteObjects(delete_input)
			}

			return lastPage
		})

	if dryRun {
		fmt.Println("(dryrun) Deleted", counter, "objects and", version_counter, "versions.")
	} else {
		fmt.Println("Deleted", counter, "objects and", version_counter, "versions.")
	}

}

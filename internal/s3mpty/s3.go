package s3mpty

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

	if result.LocationConstraint == nil {
		return "us-east-1"
	} else {
		return *result.LocationConstraint
	}

}

func NewSession() *session.Session {
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

	return sess
}

func NewClient(sess *session.Session, bucket_name string) *s3.S3 {
	region_name := aws.String(getBucketRegion(s3.New(sess), bucket_name))
	svc := s3.New(sess, aws.NewConfig().WithRegion(*region_name))
	return svc
}

func DeleteObjectsFromBucket(client *s3.S3, bucket_name string, dryRun bool) int {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket_name),
	}

	counter := 0
	err := client.ListObjectsV2Pages(input,
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
				client.DeleteObjects(delete_input)
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
	}

	return counter
}

func DeleteVersionsFromBucket(client *s3.S3, bucket_name string, dryRun bool) int {
	version_input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket_name),
	}

	version_counter := 0
	client.ListObjectVersionsPages(version_input,
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {

			delete_input := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket_name),
				Delete: &s3.Delete{Objects: []*s3.ObjectIdentifier{}},
			}
			version_counter += len(page.DeleteMarkers)
			for _, obj := range page.DeleteMarkers {
				if dryRun {
					fmt.Printf("(dryrun) delete marker: s3://%s/%s#%s\n", bucket_name, *obj.Key, *obj.VersionId)
				} else {
					delete_input.Delete.Objects = append(delete_input.Delete.Objects, &s3.ObjectIdentifier{Key: obj.Key, VersionId: obj.VersionId})
				}
			}
			version_counter += len(page.Versions)
			for _, obj := range page.Versions {
				if dryRun {
					fmt.Printf("(dryrun) delete version: s3://%s/%s#%s\n", bucket_name, *obj.Key, *obj.VersionId)
				} else {
					delete_input.Delete.Objects = append(delete_input.Delete.Objects, &s3.ObjectIdentifier{Key: obj.Key, VersionId: obj.VersionId})
				}
			}
			if !dryRun {
				client.DeleteObjects(delete_input)
			}

			return lastPage
		})

	return version_counter
}

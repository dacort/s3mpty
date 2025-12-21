package s3mpty

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3API defines the interface for S3 operations we use
type S3API interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

func getBucketRegion(ctx context.Context, svc *s3.Client, bucket_name string) string {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket_name),
	}

	result, err := svc.GetBucketLocation(ctx, input)
	if err != nil {
		log.Fatal("Error getting bucket location: ", err)
	}

	if result.LocationConstraint == "" {
		return "us-east-1"
	} else {
		return string(result.LocationConstraint)
	}

}

func NewConfig(ctx context.Context) aws.Config {
	// LoadDefaultConfig automatically loads credentials from environment,
	// shared config file, and other standard credential sources
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal("Could not load AWS config: ", err)
	}

	return cfg
}

func NewClient(ctx context.Context, cfg aws.Config, bucket_name string) *s3.Client {
	// Get the region for the bucket
	tempClient := s3.NewFromConfig(cfg)
	region_name := getBucketRegion(ctx, tempClient, bucket_name)
	
	// Create a new client with the correct region
	cfg.Region = region_name
	svc := s3.NewFromConfig(cfg)
	return svc
}

func DeleteObjectsFromBucket(ctx context.Context, client S3API, bucket_name string, prefix string, dryRun bool) int {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket_name),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	counter := 0
	var continuationToken *string
	
	for {
		input.ContinuationToken = continuationToken
		page, err := client.ListObjectsV2(ctx, input)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				fmt.Println("Bucket does not exist:", bucket_name)
			} else {
				fmt.Println("Error listing objects:", err)
			}
			break
		}

		counter += int(aws.ToInt32(page.KeyCount))

		delete_input := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket_name),
			Delete: &types.Delete{Objects: []types.ObjectIdentifier{}},
		}
		for _, obj := range page.Contents {
			if dryRun {
				fmt.Printf("(dryrun) delete: s3://%s/%s\n", bucket_name, *obj.Key)
			} else {
				delete_input.Delete.Objects = append(delete_input.Delete.Objects, types.ObjectIdentifier{Key: obj.Key})
			}

		}
		if !dryRun && len(delete_input.Delete.Objects) > 0 {
			_, err := client.DeleteObjects(ctx, delete_input)
			if err != nil {
				log.Fatal("Could not delete objects: ", err)
			}
		}
		
		if !aws.ToBool(page.IsTruncated) {
			break
		}
		continuationToken = page.NextContinuationToken
	}

	return counter
}

func DeleteVersionsFromBucket(ctx context.Context, client S3API, bucket_name string, prefix string, dryRun bool) int {
	version_input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket_name),
	}

	if prefix != "" {
		version_input.Prefix = aws.String(prefix)
	}

	version_counter := 0
	var keyMarker *string
	var versionIdMarker *string
	
	for {
		version_input.KeyMarker = keyMarker
		version_input.VersionIdMarker = versionIdMarker
		
		page, err := client.ListObjectVersions(ctx, version_input)
		if err != nil {
			log.Fatal("Could not list object versions: ", err)
		}

		delete_input := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket_name),
			Delete: &types.Delete{Objects: []types.ObjectIdentifier{}},
		}
		version_counter += len(page.DeleteMarkers)
		for _, obj := range page.DeleteMarkers {
			if dryRun {
				fmt.Printf("(dryrun) delete marker: s3://%s/%s#%s\n", bucket_name, *obj.Key, *obj.VersionId)
			} else {
				delete_input.Delete.Objects = append(delete_input.Delete.Objects, types.ObjectIdentifier{Key: obj.Key, VersionId: obj.VersionId})
			}
		}
		version_counter += len(page.Versions)
		for _, obj := range page.Versions {
			if dryRun {
				fmt.Printf("(dryrun) delete version: s3://%s/%s#%s\n", bucket_name, *obj.Key, *obj.VersionId)
			} else {
				delete_input.Delete.Objects = append(delete_input.Delete.Objects, types.ObjectIdentifier{Key: obj.Key, VersionId: obj.VersionId})
			}
		}
		if !dryRun && len(delete_input.Delete.Objects) > 0 {
			_, err := client.DeleteObjects(ctx, delete_input)
			if err != nil {
				log.Fatal("Could not delete versions: ", err)
			}
		}
		
		if !aws.ToBool(page.IsTruncated) {
			break
		}
		keyMarker = page.NextKeyMarker
		versionIdMarker = page.NextVersionIdMarker
	}

	return version_counter
}

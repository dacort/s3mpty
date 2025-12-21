package s3mpty_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	. "github.com/dacort/s3mpty/internal/s3mpty"
)

var testBucketName = "somebucket"

// Define a mock struct to be used in your unit tests of myFunc.
type mockS3Client struct {
	s3iface.S3API
	CallCount        int
	LastVersionInput *s3.ListObjectVersionsInput
	LastObjectInput  *s3.ListObjectsV2Input
}

func (m *mockS3Client) ListObjectsV2Pages(input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
	m.CallCount++
	m.LastObjectInput = input

	keycount := int64(5)
	output := s3.ListObjectsV2Output{
		KeyCount: &keycount,
		Contents: []*s3.Object{
			{Key: aws.String("file1.txt")},
			{Key: aws.String("file2.txt")},
			{Key: aws.String("file3.txt")},
			{Key: aws.String("file4.txt")},
			{Key: aws.String("file5.txt")},
		},
	}
	fn(&output, true)

	return nil
}

func (m *mockS3Client) ListObjectVersionsPages(input *s3.ListObjectVersionsInput, fn func(*s3.ListObjectVersionsOutput, bool) bool) error {
	m.CallCount++
	m.LastVersionInput = input

	output := s3.ListObjectVersionsOutput{
		DeleteMarkers: []*s3.DeleteMarkerEntry{
			{Key: aws.String("delete1.txt"), VersionId: aws.String("dv1")},
		},
		Versions: []*s3.ObjectVersion{
			{Key: aws.String("file1.txt"), VersionId: aws.String("fv1")},
			{Key: aws.String("file1.txt"), VersionId: aws.String("fv2")},
		},
	}
	fn(&output, true)

	return nil
}

func (m *mockS3Client) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	m.CallCount++

	if *input.Bucket != testBucketName {
		return nil, errors.New("Incorrect bucket")
	}

	// We don't actually use the response, so don't worry about it for now.
	return nil, nil
}

func TestDeleteObjectsFromBucketDryRun(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}

	// Verify DeleteObjectsFromBucket's functionality with a dry run
	count := DeleteObjectsFromBucket(mockSvc, testBucketName, "", true)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if mockSvc.CallCount != 1 {
		t.Errorf("expected 1 call to S3, got %v", mockSvc.CallCount)
	}
}

func TestDeleteObjectsFromBucket(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}

	// Verify myFunc's functionality
	count := DeleteObjectsFromBucket(mockSvc, testBucketName, "", false)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if mockSvc.CallCount != 2 {
		t.Errorf("expected 2 calls to S3, got %v", mockSvc.CallCount)
	}
}

func TestDeleteVersionsFromBucketDryRun(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}

	// Verify myFunc's functionality
	count := DeleteVersionsFromBucket(mockSvc, testBucketName, "", true)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if mockSvc.CallCount != 1 {
		t.Errorf("expected 1 call to S3, got %v", mockSvc.CallCount)
	}
}
func TestDeleteVersionsFromBucket(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}

	// Verify myFunc's functionality
	count := DeleteVersionsFromBucket(mockSvc, testBucketName, "", false)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if mockSvc.CallCount != 2 {
		t.Errorf("expected 2 call to S3, got %v", mockSvc.CallCount)
	}
}

func TestDeleteVersionsFromBucketWithPrefix(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}
	testPrefix := "some-prefix/"

	// Verify that prefix is passed to ListObjectVersionsInput
	count := DeleteVersionsFromBucket(mockSvc, testBucketName, testPrefix, true)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if mockSvc.LastVersionInput == nil {
		t.Error("expected LastVersionInput to be set")
	} else if mockSvc.LastVersionInput.Prefix == nil {
		t.Error("expected Prefix to be set in ListObjectVersionsInput")
	} else if *mockSvc.LastVersionInput.Prefix != testPrefix {
		t.Errorf("expected prefix to be %v, got %v", testPrefix, *mockSvc.LastVersionInput.Prefix)
	}
}

func TestDeleteObjectsFromBucketWithPrefix(t *testing.T) {
	// Setup Test
	mockSvc := &mockS3Client{}
	testPrefix := "another-prefix/"

	// Verify that prefix is passed to ListObjectsV2Input
	count := DeleteObjectsFromBucket(mockSvc, testBucketName, testPrefix, true)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if mockSvc.LastObjectInput == nil {
		t.Error("expected LastObjectInput to be set")
	} else if mockSvc.LastObjectInput.Prefix == nil {
		t.Error("expected Prefix to be set in ListObjectsV2Input")
	} else if *mockSvc.LastObjectInput.Prefix != testPrefix {
		t.Errorf("expected prefix to be %v, got %v", testPrefix, *mockSvc.LastObjectInput.Prefix)
	}
}

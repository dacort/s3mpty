package s3mpty_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	. "github.com/dacort/s3mpty/internal/s3mpty"
)

var testBucketName = "somebucket"

// mockHTTPClient implements the HTTPClient interface for testing
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Helper to create test S3 client with HTTP mocking
func newTestS3Client(t *testing.T, doFunc func(req *http.Request) (*http.Response, error)) *s3.Client {
	t.Helper()
	
	mockClient := &mockHTTPClient{DoFunc: doFunc}
	
	cfg := aws.Config{
		Region: "us-east-1",
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		}),
	}
	
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.HTTPClient = mockClient
	})
}

func TestDeleteObjectsFromBucketDryRun(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		
		// Return mocked S3 ListObjectsV2 XML response
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
    <IsTruncated>false</IsTruncated>
    <KeyCount>5</KeyCount>
    <Contents><Key>file1.txt</Key></Contents>
    <Contents><Key>file2.txt</Key></Contents>
    <Contents><Key>file3.txt</Key></Contents>
    <Contents><Key>file4.txt</Key></Contents>
    <Contents><Key>file5.txt</Key></Contents>
</ListBucketResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteObjectsFromBucket(ctx, client, testBucketName, "", true)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call to S3, got %v", callCount)
	}
}

func TestDeleteObjectsFromBucket(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		
		// Check request type based on URL path
		if strings.Contains(req.URL.RawQuery, "delete") {
			// DeleteObjects response
			body := `<?xml version="1.0" encoding="UTF-8"?>
<DeleteResult></DeleteResult>`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}
		
		// ListObjectsV2 response
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
    <IsTruncated>false</IsTruncated>
    <KeyCount>5</KeyCount>
    <Contents><Key>file1.txt</Key></Contents>
    <Contents><Key>file2.txt</Key></Contents>
    <Contents><Key>file3.txt</Key></Contents>
    <Contents><Key>file4.txt</Key></Contents>
    <Contents><Key>file5.txt</Key></Contents>
</ListBucketResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteObjectsFromBucket(ctx, client, testBucketName, "", false)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls to S3, got %v", callCount)
	}
}

func TestDeleteVersionsFromBucketDryRun(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		
		// Return mocked S3 ListObjectVersions XML response
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListVersionsResult>
    <IsTruncated>false</IsTruncated>
    <DeleteMarker>
        <Key>delete1.txt</Key>
        <VersionId>dv1</VersionId>
    </DeleteMarker>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv1</VersionId>
    </Version>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv2</VersionId>
    </Version>
</ListVersionsResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteVersionsFromBucket(ctx, client, testBucketName, "", true)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call to S3, got %v", callCount)
	}
}

func TestDeleteVersionsFromBucket(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		
		if strings.Contains(req.URL.RawQuery, "delete") {
			// DeleteObjects response
			body := `<?xml version="1.0" encoding="UTF-8"?>
<DeleteResult></DeleteResult>`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}
		
		// ListObjectVersions response
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListVersionsResult>
    <IsTruncated>false</IsTruncated>
    <DeleteMarker>
        <Key>delete1.txt</Key>
        <VersionId>dv1</VersionId>
    </DeleteMarker>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv1</VersionId>
    </Version>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv2</VersionId>
    </Version>
</ListVersionsResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteVersionsFromBucket(ctx, client, testBucketName, "", false)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls to S3, got %v", callCount)
	}
}

func TestDeleteVersionsFromBucketWithPrefix(t *testing.T) {
	ctx := context.Background()
	testPrefix := "some-prefix/"
	var capturedPrefix string
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		// Capture the prefix from the query string
		if prefix := req.URL.Query().Get("prefix"); prefix != "" {
			capturedPrefix = prefix
		}
		
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListVersionsResult>
    <IsTruncated>false</IsTruncated>
    <DeleteMarker>
        <Key>delete1.txt</Key>
        <VersionId>dv1</VersionId>
    </DeleteMarker>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv1</VersionId>
    </Version>
    <Version>
        <Key>file1.txt</Key>
        <VersionId>fv2</VersionId>
    </Version>
</ListVersionsResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteVersionsFromBucket(ctx, client, testBucketName, testPrefix, true)
	if count != 3 {
		t.Errorf("expect %v, got %v", 3, count)
	}

	if capturedPrefix == "" {
		t.Error("expected prefix to be captured")
	} else if capturedPrefix != testPrefix {
		t.Errorf("expected prefix to be %v, got %v", testPrefix, capturedPrefix)
	}
}

func TestDeleteObjectsFromBucketWithPrefix(t *testing.T) {
	ctx := context.Background()
	testPrefix := "another-prefix/"
	var capturedPrefix string
	
	client := newTestS3Client(t, func(req *http.Request) (*http.Response, error) {
		// Capture the prefix from the query string
		if prefix := req.URL.Query().Get("prefix"); prefix != "" {
			capturedPrefix = prefix
		}
		
		body := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
    <IsTruncated>false</IsTruncated>
    <KeyCount>5</KeyCount>
    <Contents><Key>file1.txt</Key></Contents>
    <Contents><Key>file2.txt</Key></Contents>
    <Contents><Key>file3.txt</Key></Contents>
    <Contents><Key>file4.txt</Key></Contents>
    <Contents><Key>file5.txt</Key></Contents>
</ListBucketResult>`
		
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	count := DeleteObjectsFromBucket(ctx, client, testBucketName, testPrefix, true)
	if count != 5 {
		t.Errorf("expect %v, got %v", 5, count)
	}

	if capturedPrefix == "" {
		t.Error("expected prefix to be captured")
	} else if capturedPrefix != testPrefix {
		t.Errorf("expected prefix to be %v, got %v", testPrefix, capturedPrefix)
	}
}

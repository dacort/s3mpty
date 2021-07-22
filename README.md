# S3 Empty

A batteries-included tool for deleting the contents of versioned S3 buckets.

## Overview

I was recently doing some testing with CDK and Amazon Managed Workflows for Apache Airflow and found myself needing to frequently clear versioned buckets.

While it's possible in the AWS Console, as well as with boto3 and the AWS CLI, I wanted a simpler, no-dependency option.

So, introducing ~~`s3mpty`~~ `rm-rs3`.

## Usage

```shell
go install github.com/dacort/s3mpty/cmd/rm-rs3@latest
```

```shell
➜ ./rm-rs3 -h
Usage: ./rm-rs3 [-dryrun] <bucket_name>
  -dryrun
        Display the operations that would be performed without actually running them.
```

Simple, as long as you are in a shell where you have authenticated to AWS (e.g. you can run `aws s3 ls` succesfully), you can run `rm-rs3 bucket_name` and it will clear out your bucket.

No prompts. No `pip installs`. Just go.

There is a `-dryrun` option if you want to see what happens.

## Testing

Want to give it a run first?

1. Create a bucket

```shell
aws s3 mb s3://somerandombucket-1234
```

2. Enable versioning

```shell
aws s3api put-bucket-versioning \
  --bucket somerandombucket-1234 \
  --versioning-configuration Status=Enabled
```

3. Upload (and delete) some data

```shell
# Change a file
echo "There was once a boy in Seattle" | \
  aws s3 cp - s3://somerandombucket-1234/file1.txt
echo "There was once a girl from Spokane" | \
  aws s3 cp - s3://somerandombucket-1234/file1.txt

# One version
echo "Doe, a deer, a female deer" | \
  aws s3 cp - s3://somerandombucket-1234/file2.txt

# Deleted file (to test if we handle DeleteMarkers)
echo "(n)Evermore" | \
  aws s3 cp - s3://somerandombucket-1234/file3.txt
aws s3 rm s3://somerandombucket-1234/file3.txt
```

Cool, now we've got some data.

```
➜ aws s3 ls s3://somerandombucket-1234/
2021-07-21 22:24:33         35 file1.txt
2021-07-21 22:26:10         27 file2.txt
```

Let's see all the versions as well.

```shell
aws s3api list-object-versions \
  --bucket somerandombucket-1234 \
  | jq -c '.Versions[] | {Key:.Key, VersionId : .VersionId}'
```

```json
{"Key":"file1.txt","VersionId":"f_ImK5Pdyed1onjCe0XkOY6FOZJ0mFpS"}
{"Key":"file1.txt","VersionId":".oHIC7wRcy5chRwsLafmcRcN8maKiwos"}
{"Key":"file2.txt","VersionId":"BqpWsB1M1ocF_Wjk49trQseu4rVtVBhU"}
{"Key":"file3.txt","VersionId":"G.fsiV42AYVwZEDfUFYxIWj9bqwj4Vvv"}
```

_And_ delete markers!

```shell
aws s3api list-object-versions \
  --bucket somerandombucket-1234 \
  | jq -c '.DeleteMarkers[] | {Key:.Key, VersionId : .VersionId}'
```

```json
{ "Key": "file3.txt", "VersionId": "VcpJgApOJTcWbnREgXPP39aAdymJnXuz" }
```

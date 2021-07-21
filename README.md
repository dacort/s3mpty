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
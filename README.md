[![Build Status](https://travis-ci.org/lonnblad/go-oc-aws-tracing.svg?branch=master)](https://travis-ci.org/lonnblad/go-oc-aws-tracing)
[![Go Report Card](https://goreportcard.com/badge/github.com/lonnblad/go-oc-aws-tracing)](https://goreportcard.com/report/github.com/lonnblad/go-oc-aws-tracing)
[![Coverage Status](https://coveralls.io/repos/github/lonnblad/go-oc-aws-tracing/badge.svg?branch=master)](https://coveralls.io/github/lonnblad/go-oc-aws-tracing?branch=master)

# go-oc-aws-tracing
go-oc-aws-tracing is a package that wraps an AWS Session, causing requests and responses to be traced by OpenCensus.

### Usage

[![Documentation](https://godoc.org/github.com/lonnblad/go-oc-aws-tracing?status.svg)](http://godoc.org/github.com/lonnblad/go-oc-aws-tracing)

``` 
import aws_trace "github.com/lonnblad/go-oc-aws-tracing"
```

``` 
cfg: = aws.NewConfig().WithRegion("us-west-2")
sess: = session.Must(session.NewSession(cfg))
sess = aws_trace.WrapSession(sess)

s3api: = s3.New(sess)
s3api.CreateBucket( & s3.CreateBucketInput {
    Bucket: aws.String("some-bucket-name"),
})
```


package aws_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	aws_trace "github.com/lonnblad/go-oc-aws-tracing"
)

// To start tracing requests, wrap the AWS session.Session by invoking
// WrapSession.
func ExampleWrapSession() {
	cfg := aws.NewConfig().WithRegion("us-west-2")
	sess := session.Must(session.NewSession(cfg))
	sess = aws_trace.WrapSession(sess)

	s3api := s3.New(sess)
	s3api.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("some-bucket-name"),
	})
}

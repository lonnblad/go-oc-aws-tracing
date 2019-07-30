package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3_api "github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"

	aws_trace "github.com/lonnblad/go-oc-aws-tracing"
)

const (
	tagAgent      = "aws.user_agent"
	tagOperation  = "aws.request.operation"
	tagService    = "aws.request.service"
	tagRegion     = "aws.request.region"
	tagMethod     = "aws.request.method"
	tagURL        = "aws.request.url"
	tagStatusCode = "aws.response.status_code"
)

func TestAWS(t *testing.T) {
	exporter := new(mockExporter)
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	cfg := aws.NewConfig().
		WithRegion("us-west-2").
		WithDisableSSL(true).
		WithCredentials(credentials.AnonymousCredentials)

	session := aws_trace.WrapSession(session.Must(session.NewSession(cfg)))

	ctx, rootSpan := trace.StartSpan(context.Background(), "test")
	s3Client := s3_api.New(session)
	_, err := s3Client.CreateBucketWithContext(ctx, &s3_api.CreateBucketInput{
		Bucket: aws.String("BUCKET"),
	})
	assert.NotNil(t, err)

	rootSpan.End()

	spans := exporter.spans
	assert.Len(t, spans, 2)
	assert.Equal(t, spans[1].TraceID, spans[0].TraceID)
	assert.Equal(t, spans[1].SpanID, spans[0].ParentSpanID)
	assert.Equal(t, 1, spans[1].ChildSpanCount)

	s := spans[0]
	assert.Contains(t, s.Attributes[tagAgent], "aws-sdk-go")
	assert.Equal(t, "CreateBucket", s.Attributes[tagOperation])
	assert.Equal(t, "s3", s.Attributes[tagService])
	assert.Equal(t, "us-west-2", s.Attributes[tagRegion])
	assert.Equal(t, "PUT", s.Attributes[tagMethod])
	assert.Equal(t, "http://s3.us-west-2.amazonaws.com/BUCKET", s.Attributes[tagURL])
	assert.Equal(t, trace.SpanKindClient, s.SpanKind)
	assert.Equal(t, "aws.s3.CreateBucket", s.Name)
	assert.Equal(t, int32(trace.StatusCodePermissionDenied), s.Status.Code)
	assert.Equal(t, "PERMISSION_DENIED", s.Status.Message)
	assert.Equal(t, int64(403), s.Attributes[tagStatusCode])
}

type mockExporter struct {
	spans []trace.SpanData
}

func (e *mockExporter) ExportSpan(s *trace.SpanData) {
	if s != nil {
		e.spans = append(e.spans, *s)
	}
}

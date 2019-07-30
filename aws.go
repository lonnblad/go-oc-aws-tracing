// Package aws implements an AWS session wrapper, causing requests and responses to be traced.
//
// The injected handler will create a new span with a name based on the called AWS service and operation.
// The span's status will be set based on the AWS response and include the following attributes:
//  - aws.user_agent
//  - aws.request.operation
//  - aws.request.service
//  - aws.request.region
//  - aws.request.method
//  - aws.request.url
//  - aws.response.status_code
//
package aws

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
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

type handlers struct{}

// WrapSession wraps an AWS Session, causing requests and responses to be traced.
func WrapSession(s *session.Session) *session.Session {
	h := &handlers{}
	s = s.Copy()
	s.Handlers.Send.PushFrontNamed(request.NamedHandler{
		Name: "github.com/lonnblad/go-oc-aws-tracing/handlers.Send",
		Fn:   h.Send,
	})
	s.Handlers.Complete.PushBackNamed(request.NamedHandler{
		Name: "github.com/lonnblad/go-oc-aws-tracing/handlers.Complete",
		Fn:   h.Complete,
	})
	return s
}

// Send is an injected handler that will be executed before the AWS request is sent.
func (h *handlers) Send(req *request.Request) {
	attributes := []trace.Attribute{
		trace.StringAttribute(tagAgent, h.awsAgent(req)),
		trace.StringAttribute(tagOperation, req.Operation.Name),
		trace.StringAttribute(tagService, req.ClientInfo.ServiceName),
		trace.StringAttribute(tagRegion, req.ClientInfo.SigningRegion),
		trace.StringAttribute(tagMethod, req.Operation.HTTPMethod),
		trace.StringAttribute(tagURL, req.HTTPRequest.URL.String()),
	}

	ctx, span := trace.StartSpan(
		req.Context(),
		"aws."+req.ClientInfo.ServiceName+"."+req.Operation.Name,
		trace.WithSpanKind(trace.SpanKindClient),
	)

	span.AddAttributes(attributes...)
	req.SetContext(ctx)
}

// Complete is an injected handler that will be executed after the AWS request is completed.
func (h *handlers) Complete(req *request.Request) {
	span := trace.FromContext(req.Context())
	if span == nil {
		return
	}

	statusCode := req.HTTPResponse.StatusCode
	span.SetStatus(ochttp.TraceStatus(statusCode, req.HTTPResponse.Status))
	span.AddAttributes(trace.Int64Attribute(tagStatusCode, int64(statusCode)))
	span.End()
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

package ocawstracing

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

const (
	tagAWSAgent     = "request.aws.agent"
	tagAWSOperation = "request.aws.operation"
	tagAWSService   = "request.aws.service"
	tagAWSRegion    = "request.aws.region"
	tagAWSRequestID = "request.aws.request_id"
)

type handlers struct{}

// WrapSession wraps a session.Session, causing requests and responses to be traced.
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

func (h *handlers) Send(req *request.Request) {
	attributes := []trace.Attribute{
		trace.StringAttribute(tagAWSRegion, req.ClientInfo.SigningRegion),
		trace.StringAttribute(tagAWSService, req.Operation.Name),
		trace.StringAttribute(tagAWSAgent, h.awsAgent(req)),
		trace.StringAttribute(tagAWSOperation, req.ClientInfo.ServiceName),
		trace.StringAttribute(tagAWSRequestID, req.RequestID),

		trace.StringAttribute(ochttp.MethodAttribute, req.Operation.HTTPMethod),
		trace.StringAttribute(ochttp.URLAttribute, req.HTTPRequest.URL.String()),
	}

	ctx, span := trace.StartSpan(
		req.Context(),
		"aws."+req.ClientInfo.ServiceName+"."+req.Operation.Name,
		trace.WithSpanKind(trace.SpanKindClient),
	)

	span.AddAttributes(attributes...)
	req.SetContext(ctx)
}

func (h *handlers) Complete(req *request.Request) {
	span := trace.FromContext(req.Context())
	if span == nil {
		return
	}

	statusCode := req.HTTPResponse.StatusCode
	span.SetStatus(ochttp.TraceStatus(statusCode, req.HTTPResponse.Status))
	span.AddAttributes(trace.Int64Attribute(ochttp.StatusCodeAttribute, int64(statusCode)))

	span.End()
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

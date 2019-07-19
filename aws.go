package ocawstracing

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.opencensus.io/trace"
)

const (
	tagAgent      = "aws.user_agent"
	tagOperation  = "aws.request.operation"
	tagService    = "aws.request.service"
	tagRegion     = "aws.request.region"
	tagRequestID  = "aws.request.request_id"
	tagMethod     = "aws.request.method"
	tagURL        = "aws.request.url"
	tagStatusCode = "aws.response.status_code"
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
		trace.StringAttribute(tagAgent, h.awsAgent(req)),

		trace.StringAttribute(tagRegion, req.ClientInfo.SigningRegion),
		trace.StringAttribute(tagService, req.ClientInfo.ServiceName),
		trace.StringAttribute(tagOperation, req.Operation.Name),
		trace.StringAttribute(tagRequestID, req.RequestID),
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

func (h *handlers) Complete(req *request.Request) {
	span := trace.FromContext(req.Context())
	if span == nil {
		return
	}

	statusCode := req.HTTPResponse.StatusCode
	span.AddAttributes(trace.Int64Attribute(tagStatusCode, int64(statusCode)))
	span.End()
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

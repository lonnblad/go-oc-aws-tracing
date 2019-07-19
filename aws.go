package ocawstracing

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

const (
	tagAWSAgent     = "aws.agent"
	tagAWSOperation = "aws.operation"
	tagAWSRegion    = "aws.region"
)

// const (
// 	HostAttribute       = "http.host"
// 	MethodAttribute     = "http.method"
// 	PathAttribute       = "http.path"
// 	URLAttribute        = "http.url"
// 	UserAgentAttribute  = "http.user_agent"
// 	StatusCodeAttribute = "http.status_code"
// )

type handlers struct {
	cfg *config
}

// WrapSession wraps a session.Session, causing requests and responses to be traced.
func WrapSession(s *session.Session, opts ...Option) *session.Session {
	cfg := new(config)
	defaults(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	h := &handlers{cfg: cfg}
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
		// 	tracer.SpanType(ext.SpanTypeHTTP),
		// 	tracer.ServiceName(h.serviceName(req)),
		// 	tracer.ResourceName(h.resourceName(req)),
		trace.StringAttribute(tagAWSAgent, h.awsAgent(req)),
		trace.StringAttribute(tagAWSOperation, h.awsOperation(req)),
		trace.StringAttribute(tagAWSRegion, h.awsRegion(req)),

		trace.StringAttribute(ochttp.MethodAttribute, req.Operation.HTTPMethod),
		trace.StringAttribute(ochttp.URLAttribute, req.HTTPRequest.URL.String()),
	}

	ctx, span := trace.StartSpan(
		req.Context(),
		h.operationName(req),
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

	// const (
	// 	HostAttribute       = "http.host"
	// 	MethodAttribute     = "http.method"
	// 	PathAttribute       = "http.path"
	// 	URLAttribute        = "http.url"
	// 	UserAgentAttribute  = "http.user_agent"
	// 	StatusCodeAttribute = "http.status_code"
	// )
	statusCode := req.HTTPResponse.StatusCode
	span.SetStatus(ochttp.TraceStatus(statusCode, req.HTTPResponse.Status))
	span.AddAttributes(trace.Int64Attribute(ochttp.StatusCodeAttribute, int64(statusCode)))

	span.End()
}

func (h *handlers) operationName(req *request.Request) string {
	return h.awsService(req) + ".command"
}

func (h *handlers) resourceName(req *request.Request) string {
	return h.awsService(req) + "." + req.Operation.Name
}

func (h *handlers) serviceName(req *request.Request) string {
	if h.cfg.serviceName != "" {
		return h.cfg.serviceName
	}
	return "aws." + h.awsService(req)
}

func (h *handlers) awsAgent(req *request.Request) string {
	if agent := req.HTTPRequest.Header.Get("User-Agent"); agent != "" {
		return agent
	}
	return "aws-sdk-go"
}

func (h *handlers) awsOperation(req *request.Request) string {
	return req.Operation.Name
}

func (h *handlers) awsRegion(req *request.Request) string {
	return req.ClientInfo.SigningRegion
}

func (h *handlers) awsService(req *request.Request) string {
	return req.ClientInfo.ServiceName
}

type config struct {
	serviceName string
}

// Option represents an option that can be passed to Dial.
type Option func(*config)

func defaults(cfg *config) {}

// WithServiceName sets the given service name for the dialled connection.
// When the service name is not explicitly set it will be inferred based on the
// request to AWS.
func WithServiceName(name string) Option {
	return func(cfg *config) {
		cfg.serviceName = name
	}
}

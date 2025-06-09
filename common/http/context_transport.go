package http_client

import (
	"net/http"

	commonCtx "github.com/dtomschitz/headless-go-client/common/context"
)

// ContextHeaderTransport adds headers from specific context keys before sending the request
type ContextHeaderTransport struct {
	Base http.RoundTripper
}

func NewContextHeaderTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &ContextHeaderTransport{Base: base}
}

func (t *ContextHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	if clientVersion := commonCtx.GetStringValue(req.Context(), commonCtx.ClientVersionKey); clientVersion != "" {
		clonedReq.Header.Set(string(commonCtx.ClientVersionKey), clientVersion)
	}
	if deviceId := commonCtx.GetStringValue(req.Context(), commonCtx.DeviceIdKey); deviceId != "" {
		clonedReq.Header.Set(string(commonCtx.DeviceIdKey), deviceId)
	}

	return t.Base.RoundTrip(clonedReq)
}

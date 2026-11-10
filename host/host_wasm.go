//go:build wasip1

package host

import (
	mail "github.com/gmomakeshop/makeshop-extensions-sdk-go/internal/hostbindings/makeshop_mail_buyer"
	connector "github.com/gmomakeshop/makeshop-extensions-sdk-go/internal/hostbindings/makeshop_runtime_host_connector"
	execution "github.com/gmomakeshop/makeshop-extensions-sdk-go/internal/hostbindings/makeshop_runtime_host_execution"
	log "github.com/gmomakeshop/makeshop-extensions-sdk-go/internal/hostbindings/makeshop_runtime_host_logging"
	uuid "github.com/gmomakeshop/makeshop-extensions-sdk-go/internal/hostbindings/makeshop_runtime_host_uuid"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Debug records a debug-level runtime log. attrsJSON must contain a JSON string.
func Debug(message, attrsJSON string) { log.Debug(message, attrsJSON) }

// Info records an info-level runtime log. attrsJSON must contain a JSON string.
func Info(message, attrsJSON string) { log.Info(message, attrsJSON) }

// Warn records a warning-level runtime log. attrsJSON must contain a JSON string.
func Warn(message, attrsJSON string) { log.Warn(message, attrsJSON) }

// Error records an error-level runtime log. attrsJSON must contain a JSON string.
func Error(message, attrsJSON string) { log.Error(message, attrsJSON) }

// Call executes a single connector request.
func Call(req ConnectorRequest) (ConnectorResponse, error) {
	return fromWitResult(connector.Call(toWitRequest(req)))
}

// CallBatch executes connector requests concurrently.
//
// Results remain in request order and contain either a response or an error for each request.
// Use sequential [Call] invocations when a request depends on an earlier result.
func CallBatch(reqs []ConnectorRequest) []CallResult {
	witReqs := make([]connector.ConnectorRequest, 0, len(reqs))
	for _, req := range reqs {
		witReqs = append(witReqs, toWitRequest(req))
	}
	results := connector.CallBatch(witReqs)
	out := make([]CallResult, 0, len(results))
	for _, result := range results {
		resp, err := fromWitResult(result)
		out = append(out, CallResult{Response: resp, Err: err})
	}
	return out
}

// Context returns metadata for the current guest extension run.
func Context() ExecutionContext {
	ctx := execution.GetContext()
	return ExecutionContext{
		RunID:           ctx.RunId,
		StartedAtUnixMs: ctx.StartedAtUnixMs,
		DeadlineUnixMs:  ctx.DeadlineUnixMs,
	}
}

// UUIDv7 returns the string representation of a version 7 UUID.
func UUIDv7() string { return uuid.GenerateUuidV7() }

// UUIDv4 returns the string representation of a version 4 UUID.
func UUIDv4() string { return uuid.GenerateUuidV4() }

// SendToBuyer reserves a mail to the buyer of the delivered order.
//
// The reservation takes effect only when the handler returns ok; a retry, dead-letter, or
// failure discards every reservation made during the run.
func SendToBuyer(m BuyerMail) error {
	result := mail.SendToBuyer(mail.BuyerMail{
		OperationKey: m.OperationKey,
		Subject:      m.Subject,
		Format:       mail.MailBodyFormat(m.Format),
		Body:         m.Body,
	})
	if result.IsErr() {
		witErr := result.Err()
		return &MailError{
			Code:    MailErrorCode(witErr.Code),
			Message: witErr.Message,
		}
	}
	return nil
}

func toWitRequest(req ConnectorRequest) connector.ConnectorRequest {
	headers := make([]connector.Header, 0, len(req.AdditionalHeaders))
	for _, header := range req.AdditionalHeaders {
		headers = append(headers, connector.Header{Name: header.Name, Value: header.Value})
	}
	timeout := witTypes.None[uint32]()
	if req.TimeoutHintMs != nil {
		timeout = witTypes.Some(*req.TimeoutHintMs)
	}
	return connector.ConnectorRequest{
		Connector:         req.Connector,
		Body:              req.Body,
		AdditionalHeaders: headers,
		TimeoutHintMs:     timeout,
	}
}

func fromWitResult(
	result witTypes.Result[connector.ConnectorResponse, connector.ConnectorError],
) (ConnectorResponse, error) {
	if result.IsErr() {
		witErr := result.Err()
		err := &ConnectorError{
			Code:    ConnectorErrorCode(witErr.Code),
			Message: witErr.Message,
		}
		if witErr.UpstreamStatus.IsSome() {
			status := witErr.UpstreamStatus.Some()
			err.UpstreamStatus = &status
		}
		return ConnectorResponse{}, err
	}

	witResp := result.Ok()
	headers := make([]Header, 0, len(witResp.Headers))
	for _, header := range witResp.Headers {
		headers = append(headers, Header{Name: header.Name, Value: header.Value})
	}
	resp := ConnectorResponse{
		Status:  witResp.Status,
		Format:  PayloadFormat(witResp.Format),
		Body:    witResp.Body,
		Headers: headers,
	}
	if witResp.UpstreamTraceId.IsSome() {
		traceID := witResp.UpstreamTraceId.Some()
		resp.UpstreamTraceID = &traceID
	}
	return resp, nil
}

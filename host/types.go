// Package host provides runtime APIs for guest extensions.
// Connector failures are represented as errors, and optional values as pointers.
package host

import "fmt"

// Header represents an HTTP header in a connector request or response.
type Header struct {
	// Name is the header name.
	Name string
	// Value is the header value.
	Value string
}

// PayloadFormat identifies the format of a connector response body.
type PayloadFormat uint8

const (
	// PayloadFormatJSON identifies JSON content.
	PayloadFormatJSON PayloadFormat = iota
	// PayloadFormatXML identifies XML content.
	PayloadFormatXML
	// PayloadFormatText identifies text content.
	PayloadFormatText
	// PayloadFormatBinary identifies binary content.
	PayloadFormatBinary
)

// ConnectorRequest represents a request passed to a connector.
//
// The deployment configuration defines the destination, credentials, and default timeout.
type ConnectorRequest struct {
	// Connector is the connector name defined in the deployment configuration.
	Connector string
	// Body is the request body. Its format is specified by the Content-Type header.
	Body []byte
	// AdditionalHeaders are HTTP headers added to those in the deployment configuration.
	AdditionalHeaders []Header
	// TimeoutHintMs requests a timeout in milliseconds. A nil value uses the configured default.
	TimeoutHintMs *uint32
}

// ConnectorResponse represents a response returned by a connector.
type ConnectorResponse struct {
	// Status is the HTTP status code.
	Status uint16
	// Format identifies the response body format.
	Format PayloadFormat
	// Body is the response body.
	Body []byte
	// Headers are the response headers.
	Headers []Header
	// UpstreamTraceID is a trace identifier, such as a request ID returned by the upstream service.
	UpstreamTraceID *string
}

// ConnectorErrorCode identifies why a connector call failed.
type ConnectorErrorCode uint8

const (
	// ConnectorErrorCodeNotFound indicates that the requested connector is not configured.
	ConnectorErrorCodeNotFound ConnectorErrorCode = iota
	// ConnectorErrorCodeTimeout indicates that the connector call timed out.
	ConnectorErrorCodeTimeout
	// ConnectorErrorCodeConnectError indicates that the request never reached the upstream
	// application, so re-sending it is safe even for a non-idempotent operation.
	ConnectorErrorCodeConnectError
	// ConnectorErrorCodeNetworkError indicates that the connection broke after the request
	// was sent, so the upstream may have processed it already.
	ConnectorErrorCodeNetworkError
	// ConnectorErrorCodePayloadInvalid indicates an invalid request payload.
	ConnectorErrorCodePayloadInvalid
	// ConnectorErrorCodeResponseTooLarge indicates that the response exceeded the size limit.
	ConnectorErrorCodeResponseTooLarge
	// ConnectorErrorCodeConcurrencyLimited indicates that the concurrency limit was reached.
	ConnectorErrorCodeConcurrencyLimited
	// ConnectorErrorCodeConnectorLimitExceeded indicates that the connector call limit was exceeded.
	ConnectorErrorCodeConnectorLimitExceeded
)

func (c ConnectorErrorCode) String() string {
	switch c {
	case ConnectorErrorCodeNotFound:
		return "not-found"
	case ConnectorErrorCodeTimeout:
		return "timeout"
	case ConnectorErrorCodeConnectError:
		return "connect-error"
	case ConnectorErrorCodeNetworkError:
		return "network-error"
	case ConnectorErrorCodePayloadInvalid:
		return "payload-invalid"
	case ConnectorErrorCodeResponseTooLarge:
		return "response-too-large"
	case ConnectorErrorCodeConcurrencyLimited:
		return "concurrency-limited"
	case ConnectorErrorCodeConnectorLimitExceeded:
		return "connector-limit-exceeded"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(c))
	}
}

// ConnectorError describes a connector call that produced no HTTP response.
//
// An HTTP response is never an error, whatever its status. A 4xx or 5xx arrives as a
// ConnectorResponse, and its Status field carries the code.
type ConnectorError struct {
	// Code identifies the failure category.
	Code ConnectorErrorCode
	// Message describes the failure.
	Message string
	// UpstreamStatus is the status the upstream returned, when its response headers arrived
	// before the call failed. That happens while reading the response: it exceeded the size
	// limit, the connection broke, or the deadline passed. It is nil whenever no response
	// reached the runtime, which is every other failure.
	//
	// It reports what happened; it does not say the call may be re-sent. Read Code for that.
	UpstreamStatus *uint16
}

func (e *ConnectorError) Error() string {
	return fmt.Sprintf("connector call failed (%s): %s", e.Code, e.Message)
}

// ExecutionContext contains metadata for a single guest extension run.
type ExecutionContext struct {
	// RunID identifies the run.
	RunID string
	// StartedAtUnixMs is the run start time as Unix time in milliseconds.
	StartedAtUnixMs uint64
	// DeadlineUnixMs is the run deadline as Unix time in milliseconds.
	DeadlineUnixMs uint64
}

// CallResult contains the result of one request passed to [CallBatch].
type CallResult struct {
	// Response contains the successful response.
	Response ConnectorResponse
	// Err contains the error when the call fails.
	Err error
}

// MailBodyFormat identifies how a [BuyerMail] body is sent.
type MailBodyFormat uint8

const (
	// MailBodyFormatText sends the body as plain text.
	MailBodyFormatText MailBodyFormat = iota
	// MailBodyFormatHTML sends the body as an HTML document. The runtime does not validate
	// or sanitize the markup.
	MailBodyFormatHTML
)

func (f MailBodyFormat) String() string {
	switch f {
	case MailBodyFormatText:
		return "text"
	case MailBodyFormatHTML:
		return "html"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(f))
	}
}

// BuyerMail is a mail reserved for the buyer of the delivered order.
//
// The runtime does not send the mail. It reserves the mail, and the platform sends it to
// the buyer only after the handler returns ok. A retry, dead-letter, or failure discards
// every reservation made during that run. A delivery succeeds at most once, so its mails
// are sent at most once. The mail is sent regardless of whether the buyer allows the shop's
// mail magazine.
type BuyerMail struct {
	// OperationKey names the mail within the delivery, such as "order-thanks", and
	// identifies the reservation in the run log. It is 1 to 128 ASCII characters: letters,
	// digits, '.', '_', ':', and '-'. It must be unique within one run.
	OperationKey string
	// Subject is the mail subject. It must contain a non-whitespace character, is limited to
	// 200 characters, and must not contain control characters or invisible formatting
	// characters such as bidirectional controls, zero width space, or byte order mark. Mail
	// variables are not expanded.
	Subject string
	// Format tells how Body is sent.
	Format MailBodyFormat
	// Body is the mail body, sent in Format. It must contain a non-whitespace character, is
	// limited in size, and must not contain control characters other than tab, line feed, and
	// carriage return.
	//
	// The platform replaces [MAIL_HEAD] in Body with the shop's common mail header and
	// [MAIL_TAIL] with its common mail footer, inserted as the shop set them. The rest of
	// Body is sent as it is.
	Body string
}

// MailErrorCode identifies why a mail reservation was rejected.
type MailErrorCode uint8

const (
	// MailErrorOperationKeyInvalid indicates that the operation key is empty, longer than
	// 128 characters, contains a character other than ASCII letters, digits, '.', '_', ':',
	// and '-', or repeats a key already reserved in this run.
	MailErrorOperationKeyInvalid MailErrorCode = iota
	// MailErrorSubjectInvalid indicates that the subject is empty or whitespace only, longer
	// than 200 characters, or contains a control character or an invisible formatting
	// character.
	MailErrorSubjectInvalid
	// MailErrorBodyInvalid indicates that the body is empty or whitespace only, too large, or
	// contains a control character other than tab, line feed, and carriage return.
	MailErrorBodyInvalid
	// MailErrorLimitExceeded indicates that this run has reached the number of mails it may
	// reserve.
	MailErrorLimitExceeded
)

func (c MailErrorCode) String() string {
	switch c {
	case MailErrorOperationKeyInvalid:
		return "operation-key-invalid"
	case MailErrorSubjectInvalid:
		return "subject-invalid"
	case MailErrorBodyInvalid:
		return "body-invalid"
	case MailErrorLimitExceeded:
		return "limit-exceeded"
	default:
		return fmt.Sprintf("unknown(%d)", uint8(c))
	}
}

// MailError describes a rejected mail reservation.
type MailError struct {
	// Code identifies the failure category.
	Code MailErrorCode
	// Message describes the failure.
	Message string
}

func (e *MailError) Error() string {
	return fmt.Sprintf("mail reservation failed (%s): %s", e.Code, e.Message)
}

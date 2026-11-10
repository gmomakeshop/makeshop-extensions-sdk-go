//go:build !wasip1

package host

import (
	"fmt"
	"maps"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

// maxMailsPerRun is the number of mails a single run may reserve via [SendToBuyer].
const maxMailsPerRun = 3

// maxMailBodyBytes is the size limit enforced on a [BuyerMail] body via [SendToBuyer].
const maxMailBodyBytes = 128 * 1024

// maxMailSubjectChars is the length limit, in characters, enforced on a [BuyerMail] subject
// via [SendToBuyer].
const maxMailSubjectChars = 200

// maxOperationKeyBytes is the length limit of [BuyerMail.OperationKey].
const maxOperationKeyBytes = 128

// validOperationKey reports whether key is 1 to [maxOperationKeyBytes] ASCII letters,
// digits, '.', '_', ':', and '-'.
func validOperationKey(key string) bool {
	if key == "" || len(key) > maxOperationKeyBytes {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		case c == '.', c == '_', c == ':', c == '-':
		default:
			return false
		}
	}
	return true
}

// LogRecord represents a runtime log captured by [TestHost].
type LogRecord struct {
	// Level is the log level.
	Level string
	// Message is the log message.
	Message string
	// AttrsJSON is a JSON string containing the log attributes.
	AttrsJSON string
}

// ConnectorHandler handles connector requests during a test.
type ConnectorHandler func(req ConnectorRequest) (ConnectorResponse, error)

type harness struct {
	mu         sync.Mutex
	connectors map[string]ConnectorHandler
	context    ExecutionContext
	logs       []LogRecord
	uuidSeq    int
	mails      []BuyerMail
	mailKeys   map[string]struct{}
}

var (
	runMu    sync.Mutex
	activeMu sync.RWMutex
	active   *harness
)

const overlappingRunPanic = "makeshop-extensions-sdk-go/host: TestHost.Run is already running; nested and concurrent runs are not supported"

// TestHost provides a host implementation for unit-testing guest extensions.
// Configure connector responses and the execution context before calling [TestHost.Run].
type TestHost struct {
	mu         sync.Mutex
	connectors map[string]ConnectorHandler
	context    ExecutionContext
	running    atomic.Bool
	mails      []BuyerMail
}

// NewTestHost returns a TestHost with no registered connectors.
func NewTestHost() *TestHost {
	return &TestHost{
		connectors: map[string]ConnectorHandler{},
		context: ExecutionContext{
			RunID:           "019bf4a4-0780-7dc2-8a95-3f65c9e518d8",
			StartedAtUnixMs: 1_767_225_600_000,
			DeadlineUnixMs:  1_767_225_601_000,
		},
	}
}

// WithConnector registers a test handler for the named connector.
func (h *TestHost) WithConnector(name string, handler ConnectorHandler) *TestHost {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running.Load() {
		panic("makeshop-extensions-sdk-go/host: a TestHost cannot be modified while TestHost.Run is in progress")
	}
	h.connectors[name] = handler
	return h
}

// WithContext sets the execution context used by the test.
func (h *TestHost) WithContext(ctx ExecutionContext) *TestHost {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running.Load() {
		panic("makeshop-extensions-sdk-go/host: a TestHost cannot be modified while TestHost.Run is in progress")
	}
	h.context = ctx
	return h
}

// Run activates the TestHost, executes body, and returns the captured logs.
//
// Run calls cannot overlap or be nested, including calls on different TestHost instances.
// Run panics when another call has not yet completed.
func (h *TestHost) Run(body func()) []LogRecord {
	if !h.running.CompareAndSwap(false, true) {
		panic(overlappingRunPanic)
	}
	if !runMu.TryLock() {
		h.running.Store(false)
		panic(overlappingRunPanic)
	}
	h.mu.Lock()
	connectors := make(map[string]ConnectorHandler, len(h.connectors))
	maps.Copy(connectors, h.connectors)
	context := h.context
	h.mu.Unlock()
	current := &harness{
		connectors: connectors,
		context:    context,
		mailKeys:   map[string]struct{}{},
	}
	activeMu.Lock()
	active = current
	activeMu.Unlock()
	defer func() {
		activeMu.Lock()
		active = nil
		activeMu.Unlock()
		runMu.Unlock()
		h.running.Store(false)
	}()
	body()
	current.mu.Lock()
	logs := append([]LogRecord(nil), current.logs...)
	mails := append([]BuyerMail(nil), current.mails...)
	current.mu.Unlock()
	h.mu.Lock()
	h.mails = mails
	h.mu.Unlock()
	return logs
}

// Mails returns the mails reserved via [SendToBuyer] during the most recent [TestHost.Run]
// call.
func (h *TestHost) Mails() []BuyerMail {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]BuyerMail(nil), h.mails...)
}

// Debug records a debug-level runtime log. attrsJSON must contain a JSON string.
func Debug(message, attrsJSON string) { record("debug", message, attrsJSON) }

// Info records an info-level runtime log. attrsJSON must contain a JSON string.
func Info(message, attrsJSON string) { record("info", message, attrsJSON) }

// Warn records a warning-level runtime log. attrsJSON must contain a JSON string.
func Warn(message, attrsJSON string) { record("warn", message, attrsJSON) }

// Error records an error-level runtime log. attrsJSON must contain a JSON string.
func Error(message, attrsJSON string) { record("error", message, attrsJSON) }

func record(level, message, attrsJSON string) {
	h := requireHarness("logging")
	h.mu.Lock()
	h.logs = append(h.logs, LogRecord{Level: level, Message: message, AttrsJSON: attrsJSON})
	h.mu.Unlock()
}

// Call executes a single connector request.
func Call(req ConnectorRequest) (ConnectorResponse, error) {
	h := requireHarness("connector")
	h.mu.Lock()
	handler, ok := h.connectors[req.Connector]
	h.mu.Unlock()
	if !ok {
		// Treat an unregistered call as a missing test setup.
		panic(fmt.Sprintf(
			"makeshop-extensions-sdk-go/host: connector %q is not registered on the TestHost; define its response with WithConnector", req.Connector))
	}
	return handler(req)
}

// CallBatch executes connector requests in input order.
// The result order and per-request success or failure match CallBatch in a Wasm build.
func CallBatch(reqs []ConnectorRequest) []CallResult {
	out := make([]CallResult, 0, len(reqs))
	for _, req := range reqs {
		resp, err := Call(req)
		out = append(out, CallResult{Response: resp, Err: err})
	}
	return out
}

// Context returns the execution context configured on the active TestHost.
func Context() ExecutionContext {
	h := requireHarness("execution")
	h.mu.Lock()
	context := h.context
	h.mu.Unlock()
	return context
}

// UUIDv7 returns a deterministic string representation of a version 7 UUID for testing.
func UUIDv7() string { return sequentialUUID(7) }

// UUIDv4 returns a deterministic string representation of a version 4 UUID for testing.
func UUIDv4() string { return sequentialUUID(4) }

// Preserve the UUID version field while using a sequence to keep test results deterministic.
func sequentialUUID(version int) string {
	h := requireHarness("uuid")
	h.mu.Lock()
	h.uuidSeq++
	sequence := h.uuidSeq
	h.mu.Unlock()
	return fmt.Sprintf("019bf4a4-0780-%d000-8a95-%012d", version, sequence)
}

// checkMailSubject reports why subject is rejected, or "" when it is valid. The subject must
// contain a non-whitespace character, be at most [maxMailSubjectChars] characters, and
// contain no undisplayable character.
func checkMailSubject(subject string) string {
	if strings.TrimSpace(subject) == "" {
		return "subject must not be empty"
	}
	if n := utf8.RuneCountInString(subject); n > maxMailSubjectChars {
		return fmt.Sprintf("subject exceeds %d characters (got %d)", maxMailSubjectChars, n)
	}
	for _, r := range subject {
		if undisplayable(r) {
			return "subject must not contain control or formatting characters"
		}
	}
	return ""
}

// undisplayable reports whether r is rejected in text shown to the buyer: a control
// character, or a formatting character that can disguise the display order or appearance
// (zero width space, left-to-right and right-to-left marks, byte order mark, and
// bidirectional embeddings, overrides, and isolates). Zero width joiner (U+200D) is allowed
// because emoji sequences need it.
func undisplayable(r rune) bool {
	switch {
	case unicode.IsControl(r):
		return true
	case r == '\u200B', r == '\u200E', r == '\u200F', r == '\uFEFF':
		return true
	case r >= '\u202A' && r <= '\u202E':
		return true
	case r >= '\u2066' && r <= '\u2069':
		return true
	}
	return false
}

// SendToBuyer reserves a mail to the buyer of the delivered order.
//
// It performs the same validation as the runtime and returns a [MailError] for a run that
// already reserved the maximum number of mails; a malformed or duplicate operation key; a
// subject that is empty or whitespace only, longer than 200 characters, or contains a
// control character or an invisible formatting character such as a bidirectional control,
// zero width space, or byte order mark; or a body that is empty (after trimming whitespace),
// too large, or contains a control character other than tab, line feed, or carriage return.
func SendToBuyer(m BuyerMail) error {
	h := requireHarness("mail")
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.mails) >= maxMailsPerRun {
		return &MailError{Code: MailErrorLimitExceeded, Message: "run already reserved the maximum number of mails"}
	}
	if !validOperationKey(m.OperationKey) {
		return &MailError{
			Code:    MailErrorOperationKeyInvalid,
			Message: "operation key must be 1 to 128 ASCII letters, digits, '.', '_', ':', and '-'",
		}
	}
	if _, exists := h.mailKeys[m.OperationKey]; exists {
		return &MailError{Code: MailErrorOperationKeyInvalid, Message: "operation key already reserved in this run"}
	}
	if message := checkMailSubject(m.Subject); message != "" {
		return &MailError{Code: MailErrorSubjectInvalid, Message: message}
	}
	if strings.TrimSpace(m.Body) == "" {
		return &MailError{Code: MailErrorBodyInvalid, Message: "body is empty"}
	}
	if len(m.Body) > maxMailBodyBytes {
		return &MailError{Code: MailErrorBodyInvalid, Message: "body exceeds the size limit"}
	}
	for _, r := range m.Body {
		if r != '\t' && r != '\n' && r != '\r' && unicode.IsControl(r) {
			return &MailError{Code: MailErrorBodyInvalid, Message: "body contains a control character"}
		}
	}

	h.mailKeys[m.OperationKey] = struct{}{}
	h.mails = append(h.mails, m)
	return nil
}

func requireHarness(capability string) *harness {
	activeMu.RLock()
	h := active
	activeMu.RUnlock()
	if h == nil {
		panic(fmt.Sprintf(
			"makeshop-extensions-sdk-go/host: %s runs only inside TestHost.Run; call the entrypoint through TestHost.Run", capability))
	}
	return h
}

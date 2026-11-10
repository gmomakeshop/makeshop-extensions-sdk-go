//go:build !wasip1

package host

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestTestHostParallelRunsDoNotMixHarnesses(t *testing.T) {
	const runs = 16
	var wg sync.WaitGroup
	errors := make(chan error, 2*runs)

	for index := range runs {
		wg.Go(func() {
			defer func() {
				if recovered := recover(); recovered != nil && !strings.Contains(fmt.Sprint(recovered), overlappingRunPanic) {
					errors <- fmt.Errorf("unexpected panic: %v", recovered)
				}
			}()
			connector := fmt.Sprintf("erp_%02d", index)
			host := NewTestHost().WithConnector(connector, func(req ConnectorRequest) (ConnectorResponse, error) {
				return ConnectorResponse{Status: 200, Body: []byte(req.Connector)}, nil
			})
			logs := host.Run(func() {
				response, err := Call(ConnectorRequest{Connector: connector})
				if err != nil {
					errors <- err
					return
				}
				if string(response.Body) != connector {
					errors <- fmt.Errorf("harness mixed: got %q want %q", response.Body, connector)
				}
				Info(connector, "{}")
			})
			if len(logs) != 1 || logs[0].Message != connector {
				errors <- fmt.Errorf("logs mixed for %s: %#v", connector, logs)
			}
		})
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

func TestTestHostNestedDifferentHostPanicsInsteadOfDeadlocking(t *testing.T) {
	outer := NewTestHost()
	inner := NewTestHost()
	var recovered any

	outer.Run(func() {
		func() {
			defer func() { recovered = recover() }()
			inner.Run(func() {})
		}()
	})

	if recovered == nil || !strings.Contains(fmt.Sprint(recovered), overlappingRunPanic) {
		t.Fatalf("nested Run must panic with overlap error, got %v", recovered)
	}
}

func TestSendToBuyerRecordsAReservation(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		err = SendToBuyer(BuyerMail{
			OperationKey: "order-thanks",
			Subject:      "Thank you for your order",
			Format:       MailBodyFormatText,
			Body:         "Thank you for your order.",
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mails := host.Mails()
	if len(mails) != 1 || mails[0].OperationKey != "order-thanks" {
		t.Fatalf("unexpected mails: %#v", mails)
	}
}

func TestSendToBuyerRejectsMalformedOperationKey(t *testing.T) {
	keys := []string{"", "order thanks", "order/thanks", "order-caf\u00e9", strings.Repeat("k", 129)}
	host := NewTestHost()
	var errs []error
	host.Run(func() {
		for _, key := range keys {
			errs = append(errs, SendToBuyer(BuyerMail{
				OperationKey: key, Subject: "subject", Format: MailBodyFormatText, Body: "hello",
			}))
		}
		errs = append(errs, SendToBuyer(BuyerMail{
			OperationKey: "Review_Request.v2:" + strings.Repeat("k", 110), Subject: "subject", Format: MailBodyFormatText, Body: "hello",
		}))
	})
	for i, key := range keys {
		var mailErr *MailError
		if !errors.As(errs[i], &mailErr) || mailErr.Code != MailErrorOperationKeyInvalid {
			t.Fatalf("key %q: expected MailErrorOperationKeyInvalid, got %v", key, errs[i])
		}
	}
	if last := errs[len(keys)]; last != nil {
		t.Fatalf("a well-formed 128-byte key must be accepted: %v", last)
	}
	if len(host.Mails()) != 1 {
		t.Fatalf("only the well-formed reservation must be recorded: %#v", host.Mails())
	}
}

func TestSendToBuyerRejectsDuplicateOperationKey(t *testing.T) {
	host := NewTestHost()
	var errs []error
	host.Run(func() {
		errs = append(errs, SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "subject", Format: MailBodyFormatText, Body: "hello",
		}))
		errs = append(errs, SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "subject", Format: MailBodyFormatText, Body: "hello",
		}))
	})
	if errs[0] != nil {
		t.Fatalf("unexpected error on first reservation: %v", errs[0])
	}
	var mailErr *MailError
	if errs[1] == nil || !errors.As(errs[1], &mailErr) || mailErr.Code != MailErrorOperationKeyInvalid {
		t.Fatalf("expected MailErrorOperationKeyInvalid, got %v", errs[1])
	}
	if len(host.Mails()) != 1 {
		t.Fatalf("duplicate reservation must not be recorded: %#v", host.Mails())
	}
}

func TestSendToBuyerRejectsControlCharacterInSubject(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		err = SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "line one\nline two", Format: MailBodyFormatText, Body: "hello",
		})
	})
	var mailErr *MailError
	if err == nil || !errors.As(err, &mailErr) || mailErr.Code != MailErrorSubjectInvalid {
		t.Fatalf("expected MailErrorSubjectInvalid, got %v", err)
	}
	if len(host.Mails()) != 0 {
		t.Fatalf("rejected reservation must not be recorded: %#v", host.Mails())
	}
}

func TestSendToBuyerRejectsEmptyBody(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		err = SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "subject", Format: MailBodyFormatText, Body: "   \n\t  ",
		})
	})
	var mailErr *MailError
	if err == nil || !errors.As(err, &mailErr) || mailErr.Code != MailErrorBodyInvalid {
		t.Fatalf("expected MailErrorBodyInvalid, got %v", err)
	}
	if len(host.Mails()) != 0 {
		t.Fatalf("rejected reservation must not be recorded: %#v", host.Mails())
	}
}

func TestSendToBuyerRejectsControlCharacterInBody(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		err = SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "subject", Format: MailBodyFormatText, Body: "hello\x00world",
		})
	})
	var mailErr *MailError
	if err == nil || !errors.As(err, &mailErr) || mailErr.Code != MailErrorBodyInvalid {
		t.Fatalf("expected MailErrorBodyInvalid, got %v", err)
	}
	if len(host.Mails()) != 0 {
		t.Fatalf("rejected reservation must not be recorded: %#v", host.Mails())
	}
}

func TestSendToBuyerAcceptsHTMLBodyAsIs(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		err = SendToBuyer(BuyerMail{
			OperationKey: "order-thanks", Subject: "subject", Format: MailBodyFormatHTML, Body: "<p>hello</p>",
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mails := host.Mails()
	if len(mails) != 1 || mails[0].Format != MailBodyFormatHTML || mails[0].Body != "<p>hello</p>" {
		t.Fatalf("unexpected mails: %#v", mails)
	}
}

func TestSendToBuyerEnforcesTheLimitPerRun(t *testing.T) {
	host := NewTestHost()
	var errs []error
	host.Run(func() {
		for index := range 4 {
			errs = append(errs, SendToBuyer(BuyerMail{
				OperationKey: fmt.Sprintf("order-thanks-%d", index),
				Subject:      "subject",
				Format:       MailBodyFormatText,
				Body:         "hello",
			}))
		}
	})
	for index := 0; index < 3; index++ {
		if errs[index] != nil {
			t.Fatalf("unexpected error on reservation %d: %v", index, errs[index])
		}
	}
	var mailErr *MailError
	if errs[3] == nil || !errors.As(errs[3], &mailErr) || mailErr.Code != MailErrorLimitExceeded {
		t.Fatalf("expected MailErrorLimitExceeded, got %v", errs[3])
	}
	if len(host.Mails()) != 3 {
		t.Fatalf("expected 3 recorded mails, got %#v", host.Mails())
	}
}

func TestTestHostStateIsSafeForGuestGoroutines(t *testing.T) {
	host := NewTestHost()
	const workers = 8
	host.Run(func() {
		var wg sync.WaitGroup
		for range workers {
			wg.Go(func() {
				Info("parallel", "{}")
				_ = UUIDv7()
			})
		}
		wg.Wait()
	})
	if logs := host.Run(func() {}); len(logs) != 0 {
		t.Fatalf("new run inherited logs: %#v", logs)
	}
}

func TestSendToBuyerSubjectFollowsTheRuntimeRules(t *testing.T) {
	longest := strings.Repeat("\u3042", 200)
	accepted := []string{
		"Thanks",
		longest,
		// Zero width joiner keeps emoji sequences intact.
		"Thanks \U0001F468\u200D\U0001F469",
	}
	rejected := []struct {
		subject string
		message string
	}{
		{"", "subject must not be empty"},
		{"   ", "subject must not be empty"},
		{"\u3000\u00A0", "subject must not be empty"},
		{longest + "a", "subject exceeds 200 characters (got 201)"},
	}
	for _, r := range []rune{
		'\x00', '\n', '\x7F', '\u0085', '\u200B', '\u200E', '\u200F', '\uFEFF', '\u202A', '\u202E', '\u2066', '\u2069',
	} {
		rejected = append(rejected, struct {
			subject string
			message string
		}{"Thanks" + string(r), "subject must not contain control or formatting characters"})
	}

	host := NewTestHost()
	var errs []error
	host.Run(func() {
		for index, subject := range accepted {
			errs = append(errs, SendToBuyer(BuyerMail{
				OperationKey: fmt.Sprintf("ok-%d", index), Subject: subject, Format: MailBodyFormatText, Body: "hello",
			}))
		}
	})
	for index, err := range errs {
		if err != nil {
			t.Fatalf("subject %q: unexpected error %v", accepted[index], err)
		}
	}

	for _, tc := range rejected {
		var err error
		NewTestHost().Run(func() {
			err = SendToBuyer(BuyerMail{
				OperationKey: "subject", Subject: tc.subject, Format: MailBodyFormatText, Body: "hello",
			})
		})
		var mailErr *MailError
		if err == nil || !errors.As(err, &mailErr) || mailErr.Code != MailErrorSubjectInvalid {
			t.Fatalf("subject %q: expected MailErrorSubjectInvalid, got %v", tc.subject, err)
		}
		if mailErr.Message != tc.message {
			t.Fatalf("subject %q: message %q, want %q", tc.subject, mailErr.Message, tc.message)
		}
	}
}

func TestSendToBuyerChecksTheLimitBeforeTheMail(t *testing.T) {
	host := NewTestHost()
	var err error
	host.Run(func() {
		for index := range 3 {
			if e := SendToBuyer(BuyerMail{
				OperationKey: fmt.Sprintf("order-thanks-%d", index), Subject: "subject", Format: MailBodyFormatText, Body: "hello",
			}); e != nil {
				t.Errorf("unexpected error on reservation %d: %v", index, e)
			}
		}
		err = SendToBuyer(BuyerMail{OperationKey: "", Subject: "", Format: MailBodyFormatText, Body: ""})
	})
	var mailErr *MailError
	if err == nil || !errors.As(err, &mailErr) || mailErr.Code != MailErrorLimitExceeded {
		t.Fatalf("expected MailErrorLimitExceeded, got %v", err)
	}
}

// Package email defines how the API sends email. The transport (SMTP today:
// Mailpit locally, Resend in production) lives in a subpackage; the rest of
// the code only sees this interface.
package email

import "context"

// Message is an email to one recipient. Text is required; HTML is optional
// and, when present, is sent as an alternative to the text.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Mailer sends email. A failure returns an apperror of kind Unavailable.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

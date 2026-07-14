package models

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxBodyLen = 4000

// MessageKind discriminates the chat payload.
type MessageKind string

const (
	MessageKindText  MessageKind = "text"
	MessageKindImage MessageKind = "image"
	MessageKindPoll  MessageKind = "poll"
)

func (k MessageKind) Valid() bool {
	switch k {
	case MessageKindText, MessageKindImage, MessageKindPoll:
		return true
	}
	return false
}

// Message is one entry in an event's chat. A date poll is a message, not a
// separate feature.
//
// Attachment and Poll are hydrated according to Kind: exactly one is set for
// image and poll respectively, and both are nil for text.
type Message struct {
	ID      uuid.UUID `json:"id"`
	EventID uuid.UUID `json:"event_id"`
	UserID  uuid.UUID `json:"user_id"`

	Kind MessageKind `json:"kind"`

	// Body carries the text for Kind text and an optional caption for image.
	// Always empty for poll, where Poll.Question is the body.
	Body string `json:"body"`

	CreatedAt time.Time `json:"created_at"`

	Attachment *Attachment `json:"attachment,omitempty"`
	Poll       *Poll       `json:"poll,omitempty"`
}

func (m Message) Validate() error {
	if m.EventID == uuid.Nil {
		return fmt.Errorf("message: event_id is required")
	}
	if m.UserID == uuid.Nil {
		return fmt.Errorf("message: user_id is required")
	}
	if utf8.RuneCountInString(m.Body) > maxBodyLen {
		return fmt.Errorf("message: body is too long")
	}

	switch m.Kind {
	case MessageKindText:
		if m.Body == "" {
			return fmt.Errorf("message: body is required in text message")
		}
		if m.Poll != nil || m.Attachment != nil {
			return fmt.Errorf("message: poll/attachment are forbidden in text message")
		}
	case MessageKindImage:
		if m.Attachment == nil {
			return fmt.Errorf("message: attachment is required in image message")
		}
		if m.Poll != nil {
			return fmt.Errorf("message: poll is forbidden in image message")
		}
	case MessageKindPoll:
		if m.Poll == nil {
			return fmt.Errorf("message: poll is required in poll message")
		}
		if m.Attachment != nil {
			return fmt.Errorf("message: attachment is forbidden in poll message")
		}
		if m.Body != "" {
			return fmt.Errorf("message: body is forbidden in poll message, the question is the body")
		}
	default:
		return fmt.Errorf("message: invalid kind")
	}

	if m.Attachment != nil {
		if err := m.Attachment.Validate(); err != nil {
			return err
		}
	}
	if m.Poll != nil {
		if err := m.Poll.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Attachment is an uploaded image belonging to a message. Only the key is stored here.
type Attachment struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`

	ObjectKey string `json:"object_key"`
	MimeType  string `json:"mime_type"`
}

// Validate is the only point where the server gets a say in what was uploaded:
// the bytes went straight to object storage, so all it sees is the key.
func (a Attachment) Validate() error {
	if a.ObjectKey == "" {
		return fmt.Errorf("attachment: object_key is required")
	}
	if !strings.HasPrefix(a.MimeType, "image/") {
		return fmt.Errorf("attachment: mime_type %q is not an image", a.MimeType)
	}
	return nil
}

package models

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func validAttachment() Attachment {
	return Attachment{
		ID:        uuid.New(),
		MessageID: uuid.New(),
		ObjectKey: "events/123/abc.jpg",
		MimeType:  "image/jpeg",
	}
}

func validTextMessage() Message {
	return Message{
		ID:      uuid.New(),
		EventID: uuid.New(),
		UserID:  uuid.New(),
		Kind:    MessageKindText,
		Body:    "see you there",
	}
}

func TestMessageValidate(t *testing.T) {
	att := validAttachment()
	poll := validPoll()

	tests := []struct {
		name    string
		mutate  func(*Message)
		wantErr bool
	}{
		{"text", func(m *Message) {}, false},
		{"text without body", func(m *Message) { m.Body = "" }, true},
		{
			"text carrying an attachment",
			func(m *Message) { m.Attachment = &att },
			true,
		},
		{
			"text carrying a poll",
			func(m *Message) { m.Poll = &poll },
			true,
		},
		{
			"body at the cap",
			func(m *Message) { m.Body = strings.Repeat("a", maxBodyLen) },
			false,
		},
		{
			"body over the cap",
			func(m *Message) { m.Body = strings.Repeat("a", maxBodyLen+1) },
			true,
		},
		{"no event", func(m *Message) { m.EventID = uuid.Nil }, true},
		{"no user", func(m *Message) { m.UserID = uuid.Nil }, true},
		{"unknown kind", func(m *Message) { m.Kind = "video" }, true},

		{
			"image with attachment",
			func(m *Message) {
				m.Kind = MessageKindImage
				m.Body = ""
				m.Attachment = &att
			},
			false,
		},
		{
			"image with a caption",
			func(m *Message) {
				m.Kind = MessageKindImage
				m.Body = "the view"
				m.Attachment = &att
			},
			false,
		},
		{
			"image without attachment",
			func(m *Message) { m.Kind = MessageKindImage; m.Body = "" },
			true,
		},
		{
			"image with a non-image attachment",
			func(m *Message) {
				bad := validAttachment()
				bad.MimeType = "application/pdf"
				m.Kind = MessageKindImage
				m.Body = ""
				m.Attachment = &bad
			},
			true,
		},

		{
			"poll",
			func(m *Message) {
				m.Kind = MessageKindPoll
				m.Body = ""
				m.Poll = &poll
			},
			false,
		},
		{
			"poll without a poll",
			func(m *Message) { m.Kind = MessageKindPoll; m.Body = "" },
			true,
		},
		{
			"poll carrying a body",
			func(m *Message) {
				m.Kind = MessageKindPoll
				m.Body = "which days can you?"
				m.Poll = &poll
			},
			true,
		},
		{
			"poll carrying an attachment",
			func(m *Message) {
				m.Kind = MessageKindPoll
				m.Body = ""
				m.Poll = &poll
				m.Attachment = &att
			},
			true,
		},
		{
			"poll carrying an invalid poll",
			func(m *Message) {
				bad := validPoll()
				bad.Question = ""
				m.Kind = MessageKindPoll
				m.Body = ""
				m.Poll = &bad
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validTextMessage()
			tt.mutate(&m)
			err := m.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestAttachmentValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Attachment)
		wantErr bool
	}{
		{"valid", func(a *Attachment) {}, false},
		{"png", func(a *Attachment) { a.MimeType = "image/png" }, false},
		{"no object key", func(a *Attachment) { a.ObjectKey = "" }, true},
		{"no mime type", func(a *Attachment) { a.MimeType = "" }, true},
		{"not an image", func(a *Attachment) { a.MimeType = "application/pdf" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := validAttachment()
			tt.mutate(&a)
			err := a.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

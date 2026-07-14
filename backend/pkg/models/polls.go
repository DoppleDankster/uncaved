package models

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxQuestionLen = 400

// Poll is a poll attached to a chat message. 2 different types:
//
//   - Date poll for a ProposedDate
//     Confirming an option writes Event.StartsAt and moves the event to
//     scheduled.
//   - Opinion poll for open questions. carry only a Label ("what are you bringing?" ->
//     food, water, gear).
type Poll struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`

	Question string `json:"question"`

	// ClosesAt is nil for a poll that stays open until confirmed.
	ClosesAt *time.Time `json:"closes_at"`

	// ConfirmedOptionID is nil until the event creator picks the winning date.
	// Only ever set on a date poll.
	ConfirmedOptionID *uuid.UUID `json:"confirmed_option_id"`

	Options []PollOption `json:"options"`
}

func (p Poll) Validate() error {
	if p.Question == "" {
		return fmt.Errorf("poll: question is required")
	}
	if utf8.RuneCountInString(p.Question) > maxQuestionLen {
		return fmt.Errorf("poll: question is too long")
	}
	if len(p.Options) < 2 {
		return fmt.Errorf("poll: at least two options are required")
	}
	for _, o := range p.Options {
		if err := o.Validate(); err != nil {
			return err
		}
	}
	// Options must be homogeneous: that is what lets a client tell a date poll
	// from an opinion poll without a stored kind.
	isDate := p.Options[0].IsDate()
	for _, o := range p.Options[1:] {
		if o.IsDate() != isDate {
			return fmt.Errorf("poll: options must be all dates or all labels")
		}
	}
	if p.ConfirmedOptionID != nil && !isDate {
		return fmt.Errorf("poll: only a date poll can be confirmed")
	}
	return nil
}

// PollOption is one choice on a poll. Exactly one of ProposedDate or Label is
// set, and every option on a poll carries the same one as its siblings — that
// homogeneity is what lets a client tell a date poll from an opinion poll
// without a stored kind.
type PollOption struct {
	ID     uuid.UUID `json:"id"`
	PollID uuid.UUID `json:"poll_id"`

	// ProposedDate is set on a date poll, nil on an opinion poll.
	ProposedDate *time.Time `json:"proposed_date"`

	// Label is the display text on an opinion poll. Empty on a date poll
	Label string `json:"label"`

	Votes []PollVote `json:"votes"`
}

// IsDate reports whether the option belongs to a date poll rather than an
// opinion poll.
func (o PollOption) IsDate() bool { return o.ProposedDate != nil }

func (o PollOption) Validate() error {
	if (o.ProposedDate != nil) == (o.Label != "") {
		return fmt.Errorf("poll option: exactly one of proposed_date or label is required")
	}
	return nil
}

// PollVote is one user's vote for one option. Keyed by (OptionID, UserID);
// voting for several options is expected
type PollVote struct {
	OptionID uuid.UUID `json:"option_id"`
	UserID   uuid.UUID `json:"user_id"`

	CreatedAt time.Time `json:"created_at"`
}

func (v PollVote) Validate() error {
	if v.OptionID == uuid.Nil {
		return fmt.Errorf("poll vote: option_id is required")
	}
	if v.UserID == uuid.Nil {
		return fmt.Errorf("poll vote: user_id is required")
	}
	return nil
}

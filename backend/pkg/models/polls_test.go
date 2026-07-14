package models

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func dateOption(at time.Time) PollOption {
	return PollOption{ID: uuid.New(), PollID: uuid.New(), ProposedDate: &at}
}

func labelOption(label string) PollOption {
	return PollOption{ID: uuid.New(), PollID: uuid.New(), Label: label}
}

// validPoll is a date poll: "which days can you?".
func validPoll() Poll {
	sat := time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC)
	sun := time.Date(2026, 8, 2, 19, 0, 0, 0, time.UTC)
	return Poll{
		ID:        uuid.New(),
		MessageID: uuid.New(),
		Question:  "Which days can you?",
		Options:   []PollOption{dateOption(sat), dateOption(sun)},
	}
}

// opinionPoll is the "what are you bringing?" shape.
func opinionPoll() Poll {
	return Poll{
		ID:        uuid.New(),
		MessageID: uuid.New(),
		Question:  "What are you bringing?",
		Options: []PollOption{
			labelOption("food"),
			labelOption("water"),
			labelOption("gear"),
		},
	}
}

func TestPollValidate(t *testing.T) {
	at := time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		poll    func() Poll
		wantErr bool
	}{
		{"date poll", validPoll, false},
		{"opinion poll", opinionPoll, false},
		{
			"no question",
			func() Poll { p := validPoll(); p.Question = ""; return p },
			true,
		},
		{
			"question at the cap",
			func() Poll { p := validPoll(); p.Question = strings.Repeat("a", maxQuestionLen); return p },
			false,
		},
		{
			"question over the cap",
			func() Poll { p := validPoll(); p.Question = strings.Repeat("a", maxQuestionLen+1); return p },
			true,
		},
		{
			"accented question at the cap",
			func() Poll { p := validPoll(); p.Question = strings.Repeat("é", maxQuestionLen); return p },
			false,
		},
		{
			"no options",
			func() Poll { p := validPoll(); p.Options = nil; return p },
			true,
		},
		{
			"single option",
			func() Poll { p := validPoll(); p.Options = p.Options[:1]; return p },
			true,
		},
		{
			"mixed date and label options",
			func() Poll {
				p := validPoll()
				p.Options = append(p.Options, labelOption("food"))
				return p
			},
			true,
		},
		{
			"option with both date and label",
			func() Poll {
				p := validPoll()
				p.Options[0].Label = "saturday"
				return p
			},
			true,
		},
		{
			"option with neither date nor label",
			func() Poll {
				p := validPoll()
				p.Options[0].ProposedDate = nil
				return p
			},
			true,
		},
		{
			"confirmed date poll",
			func() Poll {
				p := validPoll()
				id := p.Options[0].ID
				p.ConfirmedOptionID = &id
				return p
			},
			false,
		},
		{
			"confirmed opinion poll",
			func() Poll {
				p := opinionPoll()
				id := p.Options[0].ID
				p.ConfirmedOptionID = &id
				return p
			},
			true,
		},
		{
			"closes_at set",
			func() Poll { p := validPoll(); p.ClosesAt = &at; return p },
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.poll().Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestPollOptionValidate(t *testing.T) {
	at := time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		option  PollOption
		wantErr bool
	}{
		{"date only", dateOption(at), false},
		{"label only", labelOption("food"), false},
		{
			"both",
			PollOption{ID: uuid.New(), ProposedDate: &at, Label: "food"},
			true,
		},
		{"neither", PollOption{ID: uuid.New()}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.option.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestPollOptionIsDate(t *testing.T) {
	at := time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC)
	if !dateOption(at).IsDate() {
		t.Fatalf("date option should report IsDate")
	}
	if labelOption("food").IsDate() {
		t.Fatalf("label option should not report IsDate")
	}
}

func TestPollVoteValidate(t *testing.T) {
	tests := []struct {
		name    string
		vote    PollVote
		wantErr bool
	}{
		{"valid", PollVote{OptionID: uuid.New(), UserID: uuid.New()}, false},
		{"no option", PollVote{UserID: uuid.New()}, true},
		{"no user", PollVote{OptionID: uuid.New()}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vote.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

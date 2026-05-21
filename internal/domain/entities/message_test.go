package entities

import (
	"strings"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		senderID           string
		receiverID         string
		content            string
		messageType        string
		invitationGrindID  string
		invitationAccepted bool
		invitationRejected bool
		wantErr            bool
		errContains        string
	}{
		{
			name:               "creates general message",
			senderID:           " sender-1 ",
			receiverID:         " receiver-1 ",
			content:            " hello ",
			messageType:        "general",
			invitationGrindID:  "",
			invitationAccepted: false,
			invitationRejected: false,
		},
		{
			name:               "creates invitation message with grind ID",
			senderID:           "sender-1",
			receiverID:         "receiver-1",
			content:            "join my grind",
			messageType:        "invitation",
			invitationGrindID:  "grind-1",
			invitationAccepted: false,
			invitationRejected: false,
		},
		{
			name:        "rejects empty sender",
			senderID:    " ",
			receiverID:  "receiver-1",
			content:     "hello",
			messageType: "general",
			wantErr:     true,
			errContains: "senderID cannot be empty",
		},
		{
			name:        "rejects empty receiver",
			senderID:    "sender-1",
			receiverID:  " ",
			content:     "hello",
			messageType: "general",
			wantErr:     true,
			errContains: "receiverID cannot be empty",
		},
		{
			name:        "rejects empty content",
			senderID:    "sender-1",
			receiverID:  "receiver-1",
			content:     " ",
			messageType: "general",
			wantErr:     true,
			errContains: "content cannot be empty",
		},
		{
			name:        "rejects invalid type",
			senderID:    "sender-1",
			receiverID:  "receiver-1",
			content:     "hello",
			messageType: "unknown",
			wantErr:     true,
			errContains: "invalid message type",
		},
		{
			name:        "rejects invitation type without grind ID",
			senderID:    "sender-1",
			receiverID:  "receiver-1",
			content:     "join",
			messageType: "invitation",
			wantErr:     true,
			errContains: "invitationGrindID is required",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewMessage(
				tt.senderID,
				tt.receiverID,
				tt.content,
				tt.messageType,
				tt.invitationGrindID,
				tt.invitationAccepted,
				tt.invitationRejected,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if msg == nil {
				t.Fatalf("expected message, got nil")
			}
			if msg.ID == "" {
				t.Fatalf("expected non-empty message ID")
			}
			if msg.SenderID != strings.TrimSpace(tt.senderID) {
				t.Fatalf("expected normalized senderID")
			}
			if msg.ReceiverID != strings.TrimSpace(tt.receiverID) {
				t.Fatalf("expected normalized receiverID")
			}
			if msg.Content != strings.TrimSpace(tt.content) {
				t.Fatalf("expected normalized content")
			}
			if msg.Type != tt.messageType {
				t.Fatalf("expected type %q, got %q", tt.messageType, msg.Type)
			}
			if msg.Read {
				t.Fatalf("expected new message unread")
			}
			if msg.CreatedAt.IsZero() || msg.UpdatedAt.IsZero() {
				t.Fatalf("expected created/updated timestamps")
			}
			if msg.CreatedAt.Location() != time.UTC || msg.UpdatedAt.Location() != time.UTC {
				t.Fatalf("expected UTC timestamps")
			}
		})
	}
}

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		senderID   string
		receiverID string
		content    string
		wantErr    bool
	}{
		{
			name: "success", id: "msg-1", senderID: "user-1",
			receiverID: "user-2", content: "Hello!",
		},
		{name: "empty_id", id: "", senderID: "user-1", receiverID: "user-2", content: "Hi", wantErr: true},
		{name: "empty_sender", id: "msg-1", senderID: "", receiverID: "user-2", content: "Hi", wantErr: true},
		{name: "empty_receiver", id: "msg-1", senderID: "user-1", receiverID: "", content: "Hi", wantErr: true},
		{name: "empty_content", id: "msg-1", senderID: "user-1", receiverID: "user-2", content: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewMessage(tc.id, "thread-1", tc.senderID, tc.receiverID, tc.content, fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, got.Read)
				assert.Equal(t, fixedNow, got.CreatedAt)
			}
		})
	}
}

func TestMessage_MarkRead(t *testing.T) {
	msg, _ := NewMessage("msg-1", "thread-1", "user-1", "user-2", "Hello!", fixedNow)
	assert.False(t, msg.Read)
	msg.MarkRead()
	assert.True(t, msg.Read)
}

func TestNewThread(t *testing.T) {
	tests := []struct {
		name         string
		participants []string
		wantErr      bool
	}{
		{name: "success", participants: []string{"user-1", "user-2"}},
		{name: "only_one_participant", participants: []string{"user-1"}, wantErr: true},
		{name: "three_participants", participants: []string{"user-1", "user-2", "user-3"}, wantErr: true},
		{name: "same_user_twice", participants: []string{"user-1", "user-1"}, wantErr: true},
		{name: "empty_participants", participants: []string{}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewThread("thread-1", tc.participants, fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "thread-1", got.ID)
				assert.Equal(t, tc.participants, got.Participants)
			}
		})
	}
}

func TestThread_UpdateLastMessage(t *testing.T) {
	thread, _ := NewThread("thread-1", []string{"user-1", "user-2"}, fixedNow)
	later := fixedNow.AddDate(0, 0, 1)

	thread.UpdateLastMessage("How are you?", later)

	assert.Equal(t, "How are you?", thread.LastMessage)
	assert.Equal(t, later, thread.LastAt)
	assert.Equal(t, later, thread.UpdatedAt)
	assert.Equal(t, fixedNow, thread.CreatedAt) // CreatedAt unchanged
}

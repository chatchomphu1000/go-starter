package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

type messageMocks struct {
	repo  *mocks.MockMessageRepository
	clock *mocks.MockClock
	ids   *mocks.MockIDGenerator
}

func newTestMessageService(t *testing.T) (*messageService, messageMocks) {
	t.Helper()
	m := messageMocks{
		repo:  new(mocks.MockMessageRepository),
		clock: new(mocks.MockClock),
		ids:   new(mocks.MockIDGenerator),
	}
	svc := NewMessageService(m.repo, m.clock, m.ids, nopLogger()).(*messageService)
	return svc, m
}

func validThread(t *testing.T) *domain.Thread {
	t.Helper()
	thread, _ := domain.NewThread("thread-1", []string{"user-1", "user-2"}, fixedNow)
	return thread
}

func validMessage(t *testing.T) *domain.Message {
	t.Helper()
	msg, _ := domain.NewMessage("msg-1", "thread-1", "user-1", "user-2", "Hello!", fixedNow)
	return msg
}

// ─── Send ────────────────────────────────────────────────────────────────────

func TestMessageService_Send(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.SendMessageInput
		setup   func(m messageMocks)
		wantErr error
	}{
		{
			name: "success_existing_thread",
			ctx:  context.Background(),
			input: inbound.SendMessageInput{
				SenderID:   "user-1",
				ReceiverID: "user-2",
				Content:    "Hello!",
			},
			setup: func(m messageMocks) {
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().FindThread(mock.Anything, "user-1", "user-2").Return(validThread(t), nil)
				m.ids.EXPECT().New().Return("msg-1")
				m.repo.EXPECT().InsertMessage(mock.Anything, mock.Anything).Return(nil)
				m.repo.EXPECT().UpsertThread(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "success_new_thread_created",
			ctx:  context.Background(),
			input: inbound.SendMessageInput{
				SenderID:   "user-1",
				ReceiverID: "user-2",
				Content:    "Hello!",
			},
			setup: func(m messageMocks) {
				m.clock.EXPECT().Now().Return(fixedNow)
				// FindThread returns error → new thread created
				m.repo.EXPECT().FindThread(mock.Anything, "user-1", "user-2").Return(nil, domain.ErrMessageNotFound)
				// two IDs: one for thread, one for message
				m.ids.EXPECT().New().Return("thread-1").Once()
				m.ids.EXPECT().New().Return("msg-1").Once()
				m.repo.EXPECT().InsertMessage(mock.Anything, mock.Anything).Return(nil)
				m.repo.EXPECT().UpsertThread(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "context_already_cancelled",
			ctx:  cancelledCtx(),
			input: inbound.SendMessageInput{
				SenderID:   "user-1",
				ReceiverID: "user-2",
				Content:    "Hello!",
			},
			setup:   func(m messageMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "insert_message_error",
			ctx:  context.Background(),
			input: inbound.SendMessageInput{
				SenderID:   "user-1",
				ReceiverID: "user-2",
				Content:    "Hello!",
			},
			setup: func(m messageMocks) {
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().FindThread(mock.Anything, "user-1", "user-2").Return(validThread(t), nil)
				m.ids.EXPECT().New().Return("msg-1")
				m.repo.EXPECT().InsertMessage(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMessageService(t)
			tc.setup(m)

			got, err := svc.Send(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if tc.wantErr == context.Canceled {
					assert.True(t, errors.Is(err, context.Canceled))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "Hello!", got.Content)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
			m.ids.AssertExpectations(t)
		})
	}
}

// ─── GetThread ───────────────────────────────────────────────────────────────

func TestMessageService_GetThread(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		user1   string
		user2   string
		setup   func(m messageMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			user1: "user-1",
			user2: "user-2",
			setup: func(m messageMocks) {
				m.repo.EXPECT().FindThread(mock.Anything, "user-1", "user-2").Return(validThread(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			user1:   "user-1",
			user2:   "user-2",
			setup:   func(m messageMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "not_found",
			ctx:   context.Background(),
			user1: "user-1",
			user2: "user-2",
			setup: func(m messageMocks) {
				m.repo.EXPECT().FindThread(mock.Anything, "user-1", "user-2").Return(nil, domain.ErrMessageNotFound)
			},
			wantErr: domain.ErrMessageNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMessageService(t)
			tc.setup(m)

			got, err := svc.GetThread(tc.ctx, tc.user1, tc.user2)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── MarkRead ────────────────────────────────────────────────────────────────

func TestMessageService_MarkRead(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		threadID   string
		receiverID string
		setup      func(m messageMocks)
		wantErr    error
	}{
		{
			name:       "success",
			ctx:        context.Background(),
			threadID:   "thread-1",
			receiverID: "user-2",
			setup: func(m messageMocks) {
				m.repo.EXPECT().MarkThreadRead(mock.Anything, "thread-1", "user-2").Return(nil)
			},
		},
		{
			name:       "context_already_cancelled",
			ctx:        cancelledCtx(),
			threadID:   "thread-1",
			receiverID: "user-2",
			setup:      func(m messageMocks) {},
			wantErr:    context.Canceled,
		},
		{
			name:       "repo_error",
			ctx:        context.Background(),
			threadID:   "thread-1",
			receiverID: "user-2",
			setup: func(m messageMocks) {
				m.repo.EXPECT().MarkThreadRead(mock.Anything, "thread-1", "user-2").Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMessageService(t)
			tc.setup(m)

			err := svc.MarkRead(tc.ctx, tc.threadID, tc.receiverID)

			if tc.wantErr != nil {
				require.Error(t, err)
				if tc.wantErr == context.Canceled {
					assert.True(t, errors.Is(err, context.Canceled))
				}
			} else {
				require.NoError(t, err)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── ListThreads ──────────────────────────────────────────────────────────────

func TestMessageService_ListThreads(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		userID  string
		page    int
		limit   int
		setup   func(m messageMocks)
		wantErr error
	}{
		{
			name:   "success",
			ctx:    context.Background(),
			userID: "user-1",
			page:   1,
			limit:  20,
			setup: func(m messageMocks) {
				m.repo.EXPECT().FindThreadsByUser(mock.Anything, "user-1", 1, 20).Return([]*domain.Thread{validThread(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			userID:  "user-1",
			page:    1,
			limit:   20,
			setup:   func(m messageMocks) {},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMessageService(t)
			tc.setup(m)

			got, total, err := svc.ListThreads(tc.ctx, tc.userID, tc.page, tc.limit)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				_ = total
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── ListMessages ─────────────────────────────────────────────────────────────

func TestMessageService_ListMessages(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.MessageFilter
		setup   func(m messageMocks)
		wantErr error
	}{
		{
			name:   "success",
			ctx:    context.Background(),
			filter: inbound.MessageFilter{ThreadID: "thread-1", Page: 1, Limit: 50},
			setup: func(m messageMocks) {
				m.repo.EXPECT().FindMessagesByThread(mock.Anything, mock.Anything).Return([]*domain.Message{validMessage(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.MessageFilter{ThreadID: "thread-1"},
			setup:   func(m messageMocks) {},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMessageService(t)
			tc.setup(m)

			got, total, err := svc.ListMessages(tc.ctx, tc.filter)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				_ = total
			}
			m.repo.AssertExpectations(t)
		})
	}
}

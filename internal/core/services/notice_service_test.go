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

type noticeMocks struct {
	repo  *mocks.MockNoticeRepository
	clock *mocks.MockClock
	ids   *mocks.MockIDGenerator
}

func newTestNoticeService(t *testing.T) (*noticeService, noticeMocks) {
	t.Helper()
	m := noticeMocks{
		repo:  new(mocks.MockNoticeRepository),
		clock: new(mocks.MockClock),
		ids:   new(mocks.MockIDGenerator),
	}
	svc := NewNoticeService(m.repo, m.clock, m.ids, nopLogger()).(*noticeService)
	return svc, m
}

func validNotice(t *testing.T) *domain.Notice {
	t.Helper()
	n, _ := domain.NewNotice(
		"notice-1", "owner-1", "Hello", "Welcome message",
		domain.NoticeTypeGeneral, false, nil, fixedNow,
	)
	return n
}

func validCreateNoticeInput() inbound.CreateNoticeInput {
	return inbound.CreateNoticeInput{
		OwnerID: "owner-1",
		Title:   "Hello",
		Content: "Welcome message",
		Type:    "general",
		Pinned:  false,
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestNoticeService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreateNoticeInput
		setup   func(m noticeMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreateNoticeInput(),
			setup: func(m noticeMocks) {
				m.ids.EXPECT().New().Return("notice-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreateNoticeInput(),
			setup:   func(_ noticeMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "empty_title_validation_error",
			ctx:  context.Background(),
			input: inbound.CreateNoticeInput{
				OwnerID: "owner-1",
				Title:   "",
				Content: "content",
				Type:    "general",
			},
			setup: func(m noticeMocks) {
				m.ids.EXPECT().New().Return("notice-1")
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("title is required"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreateNoticeInput(),
			setup: func(m noticeMocks) {
				m.ids.EXPECT().New().Return("notice-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestNoticeService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "notice-1", got.ID)
			}
			m.repo.AssertExpectations(t)
			m.ids.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestNoticeService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m noticeMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "notice-1",
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "notice-1").Return(validNotice(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "notice-1",
			setup:   func(_ noticeMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrNoticeNotFound)
			},
			wantErr: domain.ErrNoticeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestNoticeService(t)
			tc.setup(m)

			got, err := svc.GetByID(tc.ctx, tc.id)

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

// ─── List ─────────────────────────────────────────────────────────────────────

func TestNoticeService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.NoticeFilter
		setup   func(m noticeMocks)
		wantErr error
	}{
		{
			name:   "success",
			ctx:    context.Background(),
			filter: inbound.NoticeFilter{Page: 1, Limit: 10},
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Notice{validNotice(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.NoticeFilter{},
			setup:   func(_ noticeMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.NoticeFilter{Page: 1, Limit: 10},
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestNoticeService(t)
			tc.setup(m)

			got, total, err := svc.List(tc.ctx, tc.filter)

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

// ─── Update ──────────────────────────────────────────────────────────────────

func TestNoticeService_Update(t *testing.T) {
	newTitle := "Updated Title"

	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		input   inbound.UpdateNoticeInput
		setup   func(m noticeMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "notice-1",
			ownerID: "owner-1",
			input:   inbound.UpdateNoticeInput{Title: &newTitle},
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "notice-1").Return(validNotice(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "notice-1",
			ownerID: "owner-1",
			input:   inbound.UpdateNoticeInput{Title: &newTitle},
			setup:   func(_ noticeMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "not_found",
			ctx:     context.Background(),
			id:      "missing",
			ownerID: "owner-1",
			input:   inbound.UpdateNoticeInput{Title: &newTitle},
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrNoticeNotFound)
			},
			wantErr: domain.ErrNoticeNotFound,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "notice-1",
			ownerID: "other",
			input:   inbound.UpdateNoticeInput{Title: &newTitle},
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "notice-1").Return(validNotice(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestNoticeService(t)
			tc.setup(m)

			got, err := svc.Update(tc.ctx, tc.id, tc.ownerID, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestNoticeService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m noticeMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "notice-1",
			ownerID: "owner-1",
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "notice-1").Return(validNotice(t), nil)
				m.repo.EXPECT().Delete(mock.Anything, "notice-1").Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "notice-1",
			ownerID: "owner-1",
			setup:   func(_ noticeMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "not_found",
			ctx:     context.Background(),
			id:      "missing",
			ownerID: "owner-1",
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrNoticeNotFound)
			},
			wantErr: domain.ErrNoticeNotFound,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "notice-1",
			ownerID: "other",
			setup: func(m noticeMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "notice-1").Return(validNotice(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestNoticeService(t)
			tc.setup(m)

			err := svc.Delete(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

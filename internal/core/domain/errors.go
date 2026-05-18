// Package domain contains the core business entities and sentinel errors.
package domain

import "errors"

// Sentinel errors for the domain layer.
var (
	// User errors.
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidRole        = errors.New("invalid role")
	ErrWeakPassword       = errors.New("weak password")
	ErrUserInactive       = errors.New("user inactive")

	// Room errors.
	ErrRoomNotFound     = errors.New("room not found")
	ErrRoomUnavailable  = errors.New("room unavailable")
	ErrRoomNumberExists = errors.New("room number already exists")

	// Booking errors.
	ErrBookingNotFound        = errors.New("booking not found")
	ErrBookingConflict        = errors.New("booking conflict: room already booked for this period")
	ErrInvalidBookingTransition = errors.New("invalid booking status transition")

	// Payment errors.
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrPaymentAlreadyPaid = errors.New("payment already completed")

	// Invoice errors.
	ErrInvoiceNotFound = errors.New("invoice not found")

	// Maintenance errors.
	ErrTicketNotFound = errors.New("maintenance ticket not found")

	// Notice errors.
	ErrNoticeNotFound = errors.New("notice not found")

	// Message errors.
	ErrMessageNotFound    = errors.New("message not found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to resource")
)

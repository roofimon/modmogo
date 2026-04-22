package httpadapter

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"

	"modmono/internal/customer/application"
	"modmono/internal/customer/port"
)

// --- parseLimit ---

func TestParseLimit_empty(t *testing.T) {
	got, err := parseLimit("", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 50 {
		t.Errorf("expected 50, got %d", got)
	}
}

func TestParseLimit_valid(t *testing.T) {
	got, err := parseLimit("30", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 30 {
		t.Errorf("expected 30, got %d", got)
	}
}

func TestParseLimit_clamped(t *testing.T) {
	got, err := parseLimit("200", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("expected 100 (clamped), got %d", got)
	}
}

func TestParseLimit_zero(t *testing.T) {
	_, err := parseLimit("0", 50, 100)
	if err == nil {
		t.Error("expected error for limit=0")
	}
}

func TestParseLimit_negative(t *testing.T) {
	_, err := parseLimit("-1", 50, 100)
	if err == nil {
		t.Error("expected error for negative limit")
	}
}

func TestParseLimit_nonNumeric(t *testing.T) {
	_, err := parseLimit("abc", 50, 100)
	if err == nil {
		t.Error("expected error for non-numeric input")
	}
}

// --- createErrorToHTTP ---

func TestCreateErrorToHTTP_invalidName(t *testing.T) {
	err := createErrorToHTTP(application.ErrInvalidName)
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", fe.Code)
	}
}

func TestCreateErrorToHTTP_invalidEmail(t *testing.T) {
	err := createErrorToHTTP(application.ErrInvalidEmail)
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", fe.Code)
	}
}

func TestCreateErrorToHTTP_unknownError(t *testing.T) {
	err := createErrorToHTTP(errors.New("boom"))
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", fe.Code)
	}
}

// --- idErrorToHTTP ---

func TestIDErrorToHTTP_invalidObjectID(t *testing.T) {
	err := idErrorToHTTP(application.ErrInvalidObjectID)
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", fe.Code)
	}
}

func TestIDErrorToHTTP_notFound(t *testing.T) {
	err := idErrorToHTTP(port.ErrNotFound)
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", fe.Code)
	}
}

func TestIDErrorToHTTP_unknownError(t *testing.T) {
	err := idErrorToHTTP(errors.New("boom"))
	var fe *fiber.Error
	if !errors.As(err, &fe) {
		t.Fatalf("expected *fiber.Error, got %T", err)
	}
	if fe.Code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", fe.Code)
	}
}

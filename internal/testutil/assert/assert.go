// Package assert provides test helper functions that follow testify-style naming conventions.
// All assertions call t.Fatal on failure, stopping the test immediately.
package assert

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// NoError fails the test if err is not nil.
func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Error fails the test if err is nil.
func Error(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ErrorIs fails the test if [errors.Is](err, target) is false.
func ErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("expected errors.Is(%v, %v) to be true, but got false", err, target)
	}
}

// NotErrorIs fails the test if [errors.Is](err, target) is true.
func NotErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if errors.Is(err, target) {
		t.Fatalf("expected errors.Is(%v, %v) to be false, but got true", err, target)
	}
}

// ErrorContains fails the test if err is nil or its message does not contain substr.
func ErrorContains(t *testing.T, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, but got nil", substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error containing %q, but got: %v", substr, err)
	}
}

// ErrorEqual fails the test if err is nil or its message does not exactly equal msg.
func ErrorEqual(t *testing.T, err error, msg string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error %q, got nil", msg)
	}

	if err.Error() != msg {
		t.Fatalf("expected error %q, got %q", msg, err.Error())
	}
}

// Equal fails the test if expected and actual are not equal.
func Equal[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// DeepEqual fails the test if expected and actual are not deeply equal (use for non-comparable types).
func DeepEqual(t *testing.T, expected, actual any) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// Contains fails the test if s does not contain each of the given substrings.
func Contains(t *testing.T, s string, substrings ...string) {
	t.Helper()

	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			t.Fatalf("expected %q to contain %q", s, substr)
		}
	}
}

// NotNil fails the test if v is nil.
func NotNil(t *testing.T, v any) {
	t.Helper()

	if v == nil {
		t.Fatal("expected non-nil value, got nil")
	}
}

// True fails the test if condition is false.
func True(t *testing.T, condition bool) {
	t.Helper()

	if !condition {
		t.Fatal("expected condition to be true, but it was false")
	}
}

// False fails the test if condition is true.
func False(t *testing.T, condition bool) {
	t.Helper()

	if condition {
		t.Fatal("expected condition to be false, but it was true")
	}
}

// Empty fails the test if value is not the zero value of its type.
func Empty[T comparable](t *testing.T, value T) {
	t.Helper()

	var zero T

	if value != zero {
		t.Fatalf("expected empty value, got %v", value)
	}
}

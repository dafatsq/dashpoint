package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	// AccountLockoutMaxFailures is the number of failed attempts allowed on a
	// single account before it is temporarily locked.
	AccountLockoutMaxFailures = 5
	// AccountLockoutDuration is how long an account stays locked after
	// exceeding the failure threshold.
	AccountLockoutDuration = 15 * time.Minute
)

type accountLockoutEntry struct {
	failures    int
	windowStart time.Time
	lockedUntil time.Time
}

// accountLockoutStore tracks consecutive failed authentications per account.
// It is intentionally in-memory: deployments are single-instance today, and
// the store must never outlive the process (a restart clears all locks).
type accountLockoutStore struct {
	mu      sync.Mutex
	byKey   map[string]*accountLockoutEntry
	nowFunc func() time.Time
}

func newAccountLockoutStore() *accountLockoutStore {
	return &accountLockoutStore{
		byKey:   make(map[string]*accountLockoutEntry),
		nowFunc: time.Now,
	}
}

func (s *accountLockoutStore) lockedFor(key string) (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.byKey[key]
	if !ok {
		return 0, false
	}
	if remaining := entry.lockedUntil.Sub(s.nowFunc()); remaining > 0 {
		return remaining, true
	}
	if entry.lockedUntil != (time.Time{}) && entry.lockedUntil.Before(s.nowFunc()) {
		delete(s.byKey, key)
	}
	return 0, false
}

func (s *accountLockoutStore) recordFailure(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	entry, ok := s.byKey[key]
	if !ok || now.Sub(entry.windowStart) > AccountLockoutDuration {
		entry = &accountLockoutEntry{windowStart: now}
		s.byKey[key] = entry
	}
	entry.failures++
	if entry.failures >= AccountLockoutMaxFailures {
		entry.lockedUntil = now.Add(AccountLockoutDuration)
	}
}

func (s *accountLockoutStore) reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byKey, key)
}

// AccountLockout returns middleware that temporarily locks authentication for
// a specific account (email or user id) after repeated failures, closing the
// gap where a distributed or rotating-IP attacker could grind down short PINs
// without ever tripping the per-IP limiter. Successful authentication resets
// the counter. The identifier comes from the request body; requests without a
// recognizable identifier pass through untouched (the IP limiter still
// applies).
func AccountLockout() fiber.Handler {
	store := newAccountLockoutStore()
	return func(c *fiber.Ctx) error {
		key := accountLockoutIdentifier(c)
		if key != "" {
			if remaining, locked := store.lockedFor(key); locked {
				c.Set("Retry-After", strconv.Itoa(int(remaining.Seconds())+1))
				return JSONError(
					c,
					fiber.StatusTooManyRequests,
					"ACCOUNT_LOCKED",
					fmt.Sprintf("Too many failed attempts for this account. Try again in %s.", formatLockoutDelay(remaining)),
				)
			}
		}

		if err := c.Next(); err != nil {
			// Handlers signal failures either by pre-rendering the response
			// (JSONError returns nil) or by returning *fiber.Error; count both
			// shapes so lockout cannot be bypassed.
			var fiberErr *fiber.Error
			if key != "" && errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusUnauthorized {
				store.recordFailure(key)
			}
			return err
		}

		if key == "" {
			return nil
		}
		switch c.Response().StatusCode() {
		case fiber.StatusOK, fiber.StatusCreated:
			store.reset(key)
		case fiber.StatusUnauthorized:
			store.recordFailure(key)
		}
		return nil
	}
}

// accountLockoutIdentifier extracts a stable per-account key from the login
// payloads without consuming validation responsibilities from the handlers.
func accountLockoutIdentifier(c *fiber.Ctx) string {
	var body struct {
		Email  string `json:"email"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return ""
	}
	if id := strings.TrimSpace(body.UserID); id != "" {
		return "uid:" + strings.ToLower(id)
	}
	if email := strings.TrimSpace(body.Email); email != "" {
		return "em:" + strings.ToLower(email)
	}
	return ""
}

func formatLockoutDelay(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes < 1 {
		return "less than a minute"
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}

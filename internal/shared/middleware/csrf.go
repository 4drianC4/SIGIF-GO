package middleware

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// ErrOriginNotAllowed is returned when a state-changing request comes from an
// origin that is not in the CORS allow list.
var ErrOriginNotAllowed = errors.New(errors.CodeForbidden, "origin not allowed", fiber.StatusForbidden)

// OriginGuard protects state-changing requests (POST, PUT, PATCH, DELETE)
// against CSRF now that the session travels in a cookie the browser attaches
// automatically.
//
// Browsers always send Origin on these methods, so the request is accepted
// only when its Origin (or, if missing, the origin of its Referer) is one of
// the allowed origins. Requests with neither header come from non-browser
// clients (Bruno, curl, server-to-server), which cannot be driven by a
// malicious page, and are let through. Safe methods are never blocked.
func OriginGuard(allowedOrigins []string) fiber.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if normalized := normalizeOrigin(origin); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(c *fiber.Ctx) error {
		if isSafeMethod(c.Method()) {
			return c.Next()
		}

		origin := c.Get(fiber.HeaderOrigin)
		if origin == "" {
			referer := c.Get(fiber.HeaderReferer)
			if referer == "" {
				return c.Next()
			}
			origin = refererOrigin(referer)
		}

		if _, ok := allowed[normalizeOrigin(origin)]; !ok {
			return response.Error(c, fiber.StatusForbidden, ErrOriginNotAllowed)
		}
		return c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions, fiber.MethodTrace:
		return true
	}
	return false
}

func normalizeOrigin(origin string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(origin)), "/")
}

func refererOrigin(referer string) string {
	u, err := url.Parse(referer)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

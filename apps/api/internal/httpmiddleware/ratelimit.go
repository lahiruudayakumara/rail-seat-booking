package httpmiddleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/httpx"
)

type rateWindow struct {
	count   int
	resetAt time.Time
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	var mutex sync.Mutex
	windows := make(map[string]rateWindow)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			key := r.RemoteAddr
			mutex.Lock()
			current := windows[key]
			if current.resetAt.IsZero() || !now.Before(current.resetAt) {
				current = rateWindow{resetAt: now.Add(window)}
			}
			current.count++
			windows[key] = current
			allowed := current.count <= limit
			remaining := max(0, limit-current.count)
			mutex.Unlock()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				retryAfter := max(1, int(time.Until(current.resetAt).Seconds()))
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				httpx.WriteJSON(w, http.StatusTooManyRequests, map[string]any{
					"code": "RATE_LIMITED", "message": "Too many requests. Please try again later.", "requestId": httpx.RequestID(r),
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

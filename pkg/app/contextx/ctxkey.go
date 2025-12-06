package contextx

// CtxKey is a custom type for context keys to avoid collisions.
type CtxKey string

// Exported keys for request context
const (
	CtxUserID CtxKey = "user_id"
	CtxUser   CtxKey = "user"
)

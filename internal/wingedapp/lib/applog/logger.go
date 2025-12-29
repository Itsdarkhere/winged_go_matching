package applog

import (
	"context"
)

// Logger is the interface for structured logging.
// Implementations can use logrus, zerolog, zap, etc.
//
//counterfeiter:generate . Logger
type Logger interface {
	// Debug logs at debug level
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info logs at info level
	Info(ctx context.Context, msg string, fields ...Field)
	// Warn logs at warn level
	Warn(ctx context.Context, msg string, fields ...Field)
	// Error logs at error level with an error
	Error(ctx context.Context, msg string, err error, fields ...Field)
	// Fatal logs and exits (use sparingly)
	Fatal(ctx context.Context, msg string, err error, fields ...Field)

	// WithFields returns a logger with preset fields
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// F creates a new field.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Common field constructors for consistency.
func RequestID(id string) Field     { return F("request_id", id) }
func UserID(id string) Field        { return F("user_id", id) }
func Method(m string) Field         { return F("method", m) }
func Path(p string) Field           { return F("path", p) }
func Status(s int) Field            { return F("status", s) }
func Duration(ms int64) Field       { return F("duration_ms", ms) }
func ErrorCode(c string) Field      { return F("error_code", c) }
func ErrorDetail(d string) Field    { return F("error_detail", d) }
func ErrorCategory(c string) Field  { return F("error_category", c) }
func CorrelationID(id string) Field { return F("correlation_id", id) }
func Component(c string) Field      { return F("component", c) }
func Operation(o string) Field      { return F("operation", o) }
func Resource(r string) Field       { return F("resource", r) }
func Count(n int) Field             { return F("count", n) }
func Latency(ms float64) Field      { return F("latency_ms", ms) }
func IP(addr string) Field          { return F("ip", addr) }
func UserAgent(ua string) Field     { return F("user_agent", ua) }
func TraceID(id string) Field       { return F("trace_id", id) }
func SpanID(id string) Field        { return F("span_id", id) }
func ServiceName(s string) Field    { return F("service", s) }
func Environment(e string) Field    { return F("env", e) }
func Version(v string) Field        { return F("version", v) }
func Query(q string) Field          { return F("query", q) }
func Params(p map[string]any) Field { return F("params", p) }

// Scheduling-specific field constructors for observability.
func DateInstanceID(id string) Field  { return F("date_instance_id", id) }
func MatchResultID(id string) Field   { return F("match_result_id", id) }
func InitiatorID(id string) Field     { return F("initiator_id", id) }
func ReceiverID(id string) Field      { return F("receiver_id", id) }
func SchedulingAction(a string) Field { return F("scheduling_action", a) }
func SchedulingRole(r string) Field   { return F("scheduling_role", r) }
func UIState(s string) Field          { return F("ui_state", s) }
func PreviousUIState(s string) Field  { return F("previous_ui_state", s) }
func DBStatus(s string) Field         { return F("db_status", s) }
func PreviousDBStatus(s string) Field { return F("previous_db_status", s) }
func Phase(p string) Field            { return F("phase", p) }
func ActionSuccess(ok bool) Field     { return F("action_success", ok) }
func StateTransition(from, to string) Field {
	return F("state_transition", map[string]string{"from": from, "to": to})
}
func VenueID(id string) Field      { return F("venue_id", id) }
func ProposalCount(n int) Field    { return F("proposal_count", n) }
func ScheduledTime(t string) Field { return F("scheduled_time", t) }
func DateType(dt string) Field     { return F("date_type", dt) }
func BookingStatus(s string) Field { return F("booking_status", s) }

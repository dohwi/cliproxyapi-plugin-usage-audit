package store

import (
	"encoding/json"
	"strconv"
	"time"
	"unicode/utf8"
)

// decodeLogRecordFast decodes a single trimmed JSON object line into rec,
// interning string fields directly from raw bytes via intern to avoid per-line
// string heap allocations. On any escape sequence, non-standard syntax, unknown
// field type, or non-UTC timestamp it falls back to encoding/json.Unmarshal so
// decoding semantics remain 100% identical.
func decodeLogRecordFast(data []byte, rec *LogRecord, intern map[string]string) bool {
	if len(data) < 2 || data[0] != '{' || data[len(data)-1] != '}' {
		return false
	}
	if rec == nil {
		return false
	}
	if parseLogRecordObject(data, rec, intern) {
		return true
	}
	*rec = LogRecord{}
	return json.Unmarshal(data, rec) == nil
}

func parseLogRecordObject(data []byte, rec *LogRecord, intern map[string]string) bool {
	n := len(data)
	i := 1 // skip '{'
	i = skipWS(data, i)
	if i < n && data[i] == '}' {
		return skipWS(data, i+1) == n
	}
	for {
		i = skipWS(data, i)
		if i >= n || data[i] != '"' {
			return false
		}
		keyStart := i + 1
		keyEnd, ok := scanSimpleString(data, keyStart)
		if !ok {
			return false
		}
		key := data[keyStart:keyEnd]
		i = skipWS(data, keyEnd+1)
		if i >= n || data[i] != ':' {
			return false
		}
		i = skipWS(data, i+1)
		if i >= n {
			return false
		}

		var next int
		switch string(key) {
		case "t":
			t, end, ok := parseFastUTCTime(data, i)
			if !ok {
				return false
			}
			rec.T = t
			next = end
		case "requested_at":
			t, end, ok := parseFastUTCTime(data, i)
			if !ok {
				return false
			}
			rec.RequestedAt = t
			next = end
		case "provider":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.Provider = s
			next = end
		case "executor_type":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.ExecutorType = s
			next = end
		case "model":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.Model = s
			next = end
		case "alias":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.Alias = s
			next = end
		case "api_key_id":
			// Cleared in snapshots, still parsed accurately for direct ReadAll callers.
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.APIKeyID = s
			next = end
		case "session_id":
			// High-cardinality per-conversation IDs: do not fill the low-cardinality intern map.
			s, end, ok := parseInternedString(data, i, nil)
			if !ok {
				return false
			}
			rec.SessionID = s
			next = end
		case "parent_session_id":
			s, end, ok := parseInternedString(data, i, nil)
			if !ok {
				return false
			}
			rec.ParentSessionID = s
			next = end
		case "auth_index":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.AuthIndex = s
			next = end
		case "auth_type":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.AuthType = s
			next = end
		case "source":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.Source = s
			next = end
		case "reasoning_effort":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.ReasoningEffort = s
			next = end
		case "service_tier":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.ServiceTier = s
			next = end
		case "effective_service_tier":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.EffectiveServiceTier = s
			next = end
		case "response_service_tier":
			s, end, ok := parseInternedString(data, i, intern)
			if !ok {
				return false
			}
			rec.ResponseServiceTier = s
			next = end
		case "generate":
			v, end, ok := parseFastBool(data, i)
			if !ok {
				return false
			}
			rec.Generate = v
			next = end
		case "failed":
			v, end, ok := parseFastBool(data, i)
			if !ok {
				return false
			}
			rec.Failed = v
			next = end
		case "failure_status":
			v, end, ok := parseFastInt64(data, i)
			if !ok || int64(int(v)) != v {
				return false
			}
			rec.FailureStatus = int(v)
			next = end
		case "latency_ms":
			v, end, ok := parseFastFloat64(data, i)
			if !ok {
				return false
			}
			rec.LatencyMs = v
			next = end
		case "ttft_ms":
			v, end, ok := parseFastFloat64(data, i)
			if !ok {
				return false
			}
			rec.TTFTMs = v
			next = end
		case "input_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.InputTokens = v
			next = end
		case "output_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.OutputTokens = v
			next = end
		case "reasoning_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.ReasoningTokens = v
			next = end
		case "cached_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.CachedTokens = v
			next = end
		case "cache_read_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.CacheReadTokens = v
			next = end
		case "cache_creation_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.CacheCreationTokens = v
			next = end
		case "total_tokens":
			v, end, ok := parseFastInt64(data, i)
			if !ok {
				return false
			}
			rec.TotalTokens = v
			next = end
		default:
			// Unknown key: fall back to encoding/json for exact validation.
			return false
		}

		i = skipWS(data, next)
		if i >= n {
			return false
		}
		if data[i] == ',' {
			i++
			continue
		}
		if data[i] == '}' {
			return skipWS(data, i+1) == n
		}
		return false
	}
}

func skipWS(data []byte, i int) int {
	for i < len(data) {
		switch data[i] {
		case ' ', '\t', '\r', '\n':
			i++
		default:
			return i
		}
	}
	return i
}

// scanSimpleString scans until the closing '"' and rejects escapes, control
// bytes (<0x20), or non-ASCII bytes (so invalid UTF-8 or escapes fall back to
// encoding/json for exact U+FFFD / escape handling).
func scanSimpleString(data []byte, start int) (int, bool) {
	for i := start; i < len(data); i++ {
		c := data[i]
		if c == '"' {
			return i, true
		}
		if c == '\\' || c < 0x20 {
			return 0, false
		}
		if c >= 0x80 {
			// Continue scanning multi-byte UTF-8 and validate only the string slice.
			for j := i; j < len(data); j++ {
				b := data[j]
				if b == '"' {
					if !utf8.Valid(data[start:j]) {
						return 0, false
					}
					return j, true
				}
				if b == '\\' || b < 0x20 {
					return 0, false
				}
			}
			return 0, false
		}
	}
	return 0, false
}

const maxDecodeInternEntries = 8192

func parseInternedString(data []byte, i int, intern map[string]string) (string, int, bool) {
	if i >= len(data) || data[i] != '"' {
		return "", 0, false
	}
	start := i + 1
	end, ok := scanSimpleString(data, start)
	if !ok {
		return "", 0, false
	}
	raw := data[start:end]
	if len(raw) == 0 {
		return "", end + 1, true
	}
	if intern != nil {
		if s, hit := intern[string(raw)]; hit {
			return s, end + 1, true
		}
		s := string(raw)
		if len(intern) < maxDecodeInternEntries {
			intern[s] = s
		}
		return s, end + 1, true
	}
	return string(raw), end + 1, true
}

func parseFastBool(data []byte, i int) (bool, int, bool) {
	if i+4 <= len(data) && data[i] == 't' && data[i+1] == 'r' && data[i+2] == 'u' && data[i+3] == 'e' {
		return true, i + 4, true
	}
	if i+5 <= len(data) && data[i] == 'f' && data[i+1] == 'a' && data[i+2] == 'l' && data[i+3] == 's' && data[i+4] == 'e' {
		return false, i + 5, true
	}
	return false, 0, false
}

func parseFastInt64(data []byte, i int) (int64, int, bool) {
	n := len(data)
	if i >= n {
		return 0, 0, false
	}
	neg := false
	if data[i] == '-' {
		neg = true
		i++
		if i >= n {
			return 0, 0, false
		}
	}
	if data[i] < '0' || data[i] > '9' {
		return 0, 0, false
	}
	if data[i] == '0' && i+1 < n && data[i+1] >= '0' && data[i+1] <= '9' {
		// Leading zero is invalid JSON; let fallback reject it.
		return 0, 0, false
	}
	var v uint64
	start := i
	for i < n && data[i] >= '0' && data[i] <= '9' {
		if i-start >= 18 {
			// Near overflow boundary: let fallback parse safely.
			return 0, 0, false
		}
		v = v*10 + uint64(data[i]-'0')
		i++
	}
	if i < n && (data[i] == '.' || data[i] == 'e' || data[i] == 'E') {
		return 0, 0, false
	}
	if neg {
		return -int64(v), i, true
	}
	return int64(v), i, true
}

func parseFastFloat64(data []byte, i int) (float64, int, bool) {
	n := len(data)
	start := i
	if i < n && data[i] == '-' {
		i++
	}
	digitsStart := i
	digits := 0
	for i < n && data[i] >= '0' && data[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 || (digits > 1 && data[digitsStart] == '0') {
		return 0, 0, false
	}
	isFloat := false
	if i < n && data[i] == '.' {
		isFloat = true
		i++
		frac := 0
		for i < n && data[i] >= '0' && data[i] <= '9' {
			i++
			frac++
		}
		if frac == 0 {
			return 0, 0, false
		}
	}
	if i < n && (data[i] == 'e' || data[i] == 'E') {
		isFloat = true
		i++
		if i < n && (data[i] == '+' || data[i] == '-') {
			i++
		}
		expDigits := 0
		for i < n && data[i] >= '0' && data[i] <= '9' {
			i++
			expDigits++
		}
		if expDigits == 0 {
			return 0, 0, false
		}
	}
	if !isFloat && data[start] != '-' && i-start <= 15 {
		var v uint64
		for j := start; j < i; j++ {
			v = v*10 + uint64(data[j]-'0')
		}
		return float64(v), i, true
	}
	f, err := strconv.ParseFloat(string(data[start:i]), 64)
	if err != nil {
		return 0, 0, false
	}
	return f, i, true
}

// parseFastUTCTime parses RFC3339 / RFC3339Nano UTC timestamps like
// "2026-11-24T15:08:12.123456789Z" without reflection or allocations.
// Non-UTC offsets or unusual formats return ok=false so encoding/json handles them.
func parseFastUTCTime(data []byte, i int) (time.Time, int, bool) {
	if i >= len(data) || data[i] != '"' {
		return time.Time{}, 0, false
	}
	start := i + 1
	end, ok := scanSimpleString(data, start)
	if !ok {
		return time.Time{}, 0, false
	}
	b := data[start:end]
	if len(b) < 20 || b[len(b)-1] != 'Z' {
		return time.Time{}, 0, false
	}
	if b[4] != '-' || b[7] != '-' || b[10] != 'T' || b[13] != ':' || b[16] != ':' {
		return time.Time{}, 0, false
	}
	year, ok1 := parseDigits4(b[0:4])
	month, ok2 := parseDigits2(b[5:7])
	day, ok3 := parseDigits2(b[8:10])
	hour, ok4 := parseDigits2(b[11:13])
	min, ok5 := parseDigits2(b[14:16])
	sec, ok6 := parseDigits2(b[17:19])
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 {
		return time.Time{}, 0, false
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || hour > 23 || min > 59 || sec > 59 {
		return time.Time{}, 0, false
	}
	nsec := 0
	if len(b) > 20 {
		if b[19] != '.' {
			return time.Time{}, 0, false
		}
		frac := b[20 : len(b)-1]
		if len(frac) == 0 || len(frac) > 9 {
			return time.Time{}, 0, false
		}
		mult := 100_000_000
		for _, c := range frac {
			if c < '0' || c > '9' {
				return time.Time{}, 0, false
			}
			nsec += int(c-'0') * mult
			mult /= 10
		}
	}
	t := time.Date(year, time.Month(month), day, hour, min, sec, nsec, time.UTC)
	// Reject invalid calendar dates like Feb 30 (time.Date normalizes them).
	if int(t.Month()) != month || t.Day() != day {
		return time.Time{}, 0, false
	}
	return t, end + 1, true
}

func parseDigits2(b []byte) (int, bool) {
	if b[0] < '0' || b[0] > '9' || b[1] < '0' || b[1] > '9' {
		return 0, false
	}
	return int(b[0]-'0')*10 + int(b[1]-'0'), true
}

func parseDigits4(b []byte) (int, bool) {
	hi, ok1 := parseDigits2(b[0:2])
	lo, ok2 := parseDigits2(b[2:4])
	if !ok1 || !ok2 {
		return 0, false
	}
	return hi*100 + lo, true
}

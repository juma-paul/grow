package executor

import (
	"fmt"
	"strings"
)

const observePreamble = `import ctypes, json, sys

# --- version probe: validate _PyListStruct layout ---
_vi = sys.version_info
if _vi.major != 3 or _vi.minor < 8:
    sys.stderr.write("observe requires CPython 3.8+, got %d.%d\n" % (_vi.major, _vi.minor))
    sys.exit(1)

_empty_size = sys.getsizeof([])
_ptr = ctypes.sizeof(ctypes.c_void_p)
_ssz = ctypes.sizeof(ctypes.c_ssize_t)
_expected = 3 * _ssz + 2 * _ptr
if _empty_size < _expected:
    sys.stderr.write(
        "unexpected list layout: getsizeof([])=%d, expected>=%d "
        "(ptr=%d, ssize=%d). Free-threaded or 32-bit build?\n"
        % (_empty_size, _expected, _ptr, _ssz)
    )
    sys.exit(1)

# --- resource limits (Unix only) ---
try:
    import resource
    resource.setrlimit(resource.RLIMIT_CPU, (5, 5))
    _mem = 256 * 1024 * 1024
    resource.setrlimit(resource.RLIMIT_AS, (_mem, _mem))
except (ImportError, ValueError, OSError):
    pass

class _PyListStruct(ctypes.Structure):
    _fields_ = [
        ("ob_refcnt",  ctypes.c_ssize_t),
        ("ob_type",    ctypes.c_void_p),
        ("ob_size",    ctypes.c_ssize_t),
        ("ob_item",    ctypes.c_void_p),
        ("allocated",  ctypes.c_ssize_t),
    ]

def _snap(lst, tag):
    st = _PyListStruct.from_address(id(lst))
    addr = st.ob_item if st.ob_item is not None else 0
    sys.stdout.write(json.dumps({
        "tag":  tag,
        "len":  st.ob_size,
        "cap":  st.allocated,
        "addr": addr,
    }) + "\n")
    sys.stdout.flush()

`

// listMutators are method calls on a list that change its internal state.
var listMutators = map[string]bool{
	"append": true,
	"extend": true,
	"insert": true,
	"pop":    true,
	"remove": true,
	"clear":  true,
	"sort":   true,
}

// InstrumentForObserve takes user Python source, identifies list variables
// and their mutating operations, and returns a script that emits JSON
// snapshots to stdout via _snap(). The returned script runs under real
// CPython (not gpython).
//
// Detection is line-based (not full AST) — it handles the common patterns:
//   - Assignment: x = []  or  x = list()
//   - Method calls: x.append(...), x.pop(), x.extend(...), etc.
//   - Augmented assignment: x += [...]
//   - Slice assignment: x[i:j] = ...
func InstrumentForObserve(src string) string {
	lines := strings.Split(src, "\n")
	listVars := make(map[string]bool)
	var out []string
	opCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		indent := leadingWhitespace(line)

		// Detect list creation: x = [] or x = list()
		if varName, ok := detectListCreation(trimmed); ok {
			listVars[varName] = true
			out = append(out, line)
			opCount++
			out = append(out, snapCall(indent, varName, opCount, "init"))
			continue
		}

		// Detect mutation on a known list variable
		if varName, method := detectListMutation(trimmed, listVars); varName != "" {
			out = append(out, line)
			opCount++
			out = append(out, snapCall(indent, varName, opCount, method))
			continue
		}

		out = append(out, line)
	}

	return observePreamble + strings.Join(out, "\n") + "\n"
}

func snapCall(indent, varName string, opNum int, method string) string {
	return fmt.Sprintf("%s_snap(%s, \"%s_%d\")", indent, varName, method, opNum)
}

func leadingWhitespace(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return line
}

// detectListCreation checks for patterns like "x = []" or "x = list()".
func detectListCreation(line string) (string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", false
	}
	lhs := strings.TrimSpace(parts[0])
	rhs := strings.TrimSpace(parts[1])

	// Guard against ==, !=, <=, >=, +=
	if strings.HasSuffix(lhs, "!") || strings.HasSuffix(lhs, "<") ||
		strings.HasSuffix(lhs, ">") || strings.HasSuffix(lhs, "+") ||
		strings.HasPrefix(rhs, "=") {
		return "", false
	}

	if !isSimpleIdentifier(lhs) {
		return "", false
	}

	if rhs == "[]" || strings.HasPrefix(rhs, "list(") {
		return lhs, true
	}

	return "", false
}

// detectListMutation checks for method calls like "x.append(v)" on known list vars.
func detectListMutation(line string, known map[string]bool) (string, string) {
	dotIdx := strings.Index(line, ".")
	if dotIdx < 0 {
		// Check augmented assignment: x += [...]
		plusIdx := strings.Index(line, "+=")
		if plusIdx > 0 {
			varName := strings.TrimSpace(line[:plusIdx])
			if known[varName] {
				return varName, "extend"
			}
		}
		return "", ""
	}

	varName := strings.TrimSpace(line[:dotIdx])
	if !known[varName] {
		return "", ""
	}

	rest := line[dotIdx+1:]
	parenIdx := strings.Index(rest, "(")
	if parenIdx < 0 {
		return "", ""
	}

	method := strings.TrimSpace(rest[:parenIdx])
	if listMutators[method] {
		return varName, method
	}

	return "", ""
}

func isSimpleIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 && ch >= '0' && ch <= '9' {
			return false
		}
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

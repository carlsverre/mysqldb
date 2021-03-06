package escape

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/carlsverre/go-libmysql/libmysql/bridge"
)

// Escapes the provided value such that it is ready to be inserted directly into a query
func Escape(val driver.Value) (out string, err error) {
	switch val := val.(type) {
	case int:
		out = strconv.FormatInt(int64(val), 10)
	case int32:
		out = strconv.FormatInt(int64(val), 10)
	case int64:
		out = strconv.FormatInt(val, 10)
	case float64:
		out = strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		out = strconv.FormatBool(val)
	case string:
		out = escapeString(val)
	case time.Time:
		out = escapeTime(val)
	default:
		if val == nil {
			out = "NULL"
		} else {
			err = errors.New(fmt.Sprintf("Cannot escape value of type %s", reflect.TypeOf(val)))
		}
	}

	return out, err
}

func EscapeQuery(query string, args []driver.Value) (string, error) {
	var buf bytes.Buffer

	end := len(query)
	argsLen := len(args)
	argIndex := 0

	for i := 0; i < end; i++ {
		lasti := i
		for i < end && query[i] != '%' {
			i++
		}
		if i > lasti {
			buf.WriteString(query[lasti:i])
		}
		if i >= end {
			break
		}

		// process the flag
		i++
		switch query[i] {
		case '%':
			// escape % with %%
			buf.WriteRune('%')
		case 's':
			if argIndex >= argsLen {
				return "", errors.New("Not enough arguments provided")
			}
			out, err := Escape(args[argIndex])
			if err != nil {
				return "", err
			}
			argIndex++
			buf.WriteString(out)
		default:
			return "", fmt.Errorf("Invalid format string %%%c", query[i])
		}
	}

	if argIndex != argsLen {
		return "", errors.New("Too many arguments provided")
	}

	return buf.String(), nil
}

func escapeString(val string) string {
	out := bridge.EscapeString(val)
	return "'" + out + "'"
}

func escapeTime(val time.Time) string {
	var out string

	if val.IsZero() {
		out = "'0000-00-00'"
	} else {
		out = escapeString(val.Format(time.RFC3339))
	}

	return out
}

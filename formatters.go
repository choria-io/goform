// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package goform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/tidwall/gjson"
)

func (f *formatter) getDataItem(data json.RawMessage, item string) (gjson.Result, error) {
	var sum bool
	if strings.HasPrefix(item, "report.summary.") {
		// we place items in with . so gjson needs us to escape that dot. So report.summary.x.y becomes report.summary.x\.y
		item = "report.summary." + strings.ReplaceAll(strings.TrimPrefix(item, "report.summary."), ".", "\\.")
		sum = true
	}

	val := gjson.GetBytes(data, item)

	if (!val.Exists() || val.Type == gjson.Null) && sum {
		val = gjson.Parse("0")
	}

	return val, nil
}

func formatDataItem(val gjson.Result, format string) (string, error) {
	if !val.Exists() || val.Type == gjson.Null {
		return "?" + strings.Repeat(" ", len(format)-1), nil
	}

	switch val.Type {
	case gjson.String:
		return formatString(val.String(), format)

	case gjson.Number:
		if strings.Contains(format, ".") {
			return formatNumber(val.Float(), format)
		} else {
			return formatNumber(val.Int(), format)
		}

	case gjson.True, gjson.False:
		return formatString(val.String(), format)

	default:
		return "", fmt.Errorf("invalid type %s", val.Type.String())
	}
}

func formatString(v string, format string) (string, error) {
	if len(format) == 0 {
		return "", fmt.Errorf("invalid format: %s", format)
	}

	// only up to first new line
	if strings.Contains(v, "\n") {
		parts := strings.Split(v, "\n")
		v = parts[0]
	}

	makeSuf := func(v string) string {
		suf := string(format[len(format)-1])
		if strings.ContainsAny(suf, "<>|") {
			return v
		}

		if len(v) >= len(format) {
			v = v[0 : len(format)-1]
		}

		return v + suf
	}

	lv := len(v)
	lf := len(format)

	// its already the full length no formatting needed
	if lv == lf {
		return makeSuf(v), nil
	}

	// if its longer we just take up to the length
	if lv > lf {
		return makeSuf(v[0:lf]), nil
	}

	switch format[1] {
	case '|':
		v = makeSuf(v)
		lv = len(v)

		leftPad := strings.Repeat(" ", (lf-lv)/2)
		rightPad := strings.Repeat(" ", lf-len(leftPad)-lv)
		return leftPad + v + rightPad, nil

	case '>':
		v = makeSuf(v)
		lv = len(v)

		return strings.Repeat(" ", lf-lv) + v, nil

	case '<':
		v = makeSuf(v)
		lv = len(v)

		return v + strings.Repeat(" ", lf-lv), nil

	default:
		return "", fmt.Errorf("invalid string format: %s", format)
	}
}

func formatNumber(v any, format string) (string, error) {
	var val string
	var err error

	switch {
	case format[1] == '0' || format[1] == '#' || strings.Contains(format, "."):
		val, err = formatFloat(v, format)
	case format[1] == ',':
		val, err = formatCommaNumber(v, format)
	case format[1] == 'B':
		val, err = formatBytesNumber(v, format)
	default:
		return "", fmt.Errorf("unknown number format: %v", format)
	}
	if err != nil {
		return "", err
	}

	if len(val) == 0 {
		return strings.Repeat(" ", len(format)), nil
	}

	if len(val) > len(format) {
		return strings.Repeat("#", len(format)), nil
	}

	if len(val) < len(format) {
		if format[1] == '0' {
			val = strings.Repeat("0", len(format)-len(val)) + val
		} else {
			val = val + strings.Repeat(" ", len(format)-len(val))
		}
	}

	return val, nil
}

func formatFloat(v any, format string) (string, error) {
	parts := strings.Split(format, ".")
	lp := len(parts)

	var sfmt string

	switch {
	case lp == 1:
		sfmt = "%d"
	case lp == 2:
		sfmt = fmt.Sprintf("%%%d.%df", len(parts[0]), len(parts[1]))
	default:
		return "", fmt.Errorf("invalid number format: %s", format)
	}

	return fmt.Sprintf(sfmt, v), nil
}

func formatCommaNumber(v any, format string) (string, error) {
	switch nv := v.(type) {
	case float64:
		return humanize.Commaf(nv), nil
	case float32:
		return humanize.Commaf(float64(nv)), nil
	case int:
		return humanize.Comma(int64(nv)), nil
	case int8:
		return humanize.Comma(int64(nv)), nil
	case int16:
		return humanize.Comma(int64(nv)), nil
	case int32:
		return humanize.Comma(int64(nv)), nil
	case int64:
		return humanize.Comma(int64(nv)), nil
	default:
		return "", fmt.Errorf("do not know how to handle value %v", v)
	}
}

func formatBytesNumber(v any, format string) (string, error) {
	switch nv := v.(type) {
	case float64:
		return humanize.IBytes(uint64(nv)), nil
	case int:
		return humanize.IBytes(uint64(nv)), nil
	case int8:
		return humanize.IBytes(uint64(nv)), nil
	case int16:
		return humanize.IBytes(uint64(nv)), nil
	case int32:
		return humanize.IBytes(uint64(nv)), nil
	case int64:
		return humanize.IBytes(uint64(nv)), nil
	default:
		return "", fmt.Errorf("do not know how to handle value %v", v)
	}
}

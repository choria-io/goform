// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package goform

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func TestFormatNumber(t *testing.T) {
	var cases = []struct {
		value     any
		format    string
		expect    string
		expectErr string
	}{
		// formatFloat
		{1, "@##", "1  ", ""},
		{1.1, "@#.##", "1.10 ", ""},
		{1.0, "@.##", "1.00", ""},
		{1.0, "@##.##", "1.00  ", ""},
		{int64(123456), "@,########", "123,456   ", ""},
		{float64(123456), "@,########", "123,456   ", ""},
		// formatCommaNumber - various types
		{123456.12, "@,########", "123,456.12", ""},
		{float32(1234), "@,########", "1,234     ", ""},
		{int8(100), "@,########", "100       ", ""},
		{int16(12345), "@,########", "12,345    ", ""},
		{int32(123456), "@,########", "123,456   ", ""},
		{int64(123456), "@,########", "123,456   ", ""},
		// formatBytesNumber - various types
		{123456, "@B#####", "121 KiB", ""},
		{float64(1048576), "@B########", "1.0 MiB   ", ""},
		{int8(100), "@B####", "100 B ", ""},
		{int16(1024), "@B########", "1.0 KiB   ", ""},
		{int32(1048576), "@B########", "1.0 MiB   ", ""},
		{int64(1048576), "@B########", "1.0 MiB   ", ""},
		// misc
		{123456.123, "@,########", "##########", ""},
		{1.1, "@0.##", "01.10", ""},
		// unknown number format
		{1, "@X####", "", "unknown number format: @X####"},
		// unknown type in comma format
		{"hello", "@,####", "", "do not know how to handle value hello"},
		// unknown type in bytes format
		{"hello", "@B####", "", "do not know how to handle value hello"},
		// invalid float format with too many dots
		{1.0, "@#.#.#", "", "invalid number format: @#.#.#"},
	}

	for _, tc := range cases {
		res, err := formatNumber(tc.value, tc.format)
		if tc.expectErr == "" && err != nil {
			t.Fatalf("Did not expect an error, got: %v", err)
		}
		if tc.expectErr != "" && (err == nil || err.Error() != tc.expectErr) {
			t.Fatalf("expected err %s: %v", tc.expectErr, err)
		}
		if res != tc.expect {
			t.Fatalf("expected '%v' for format '%v' got '%v'", tc.expect, tc.format, res)
		}
	}
}

func TestFormatString(t *testing.T) {
	var cases = []struct {
		value     string
		format    string
		expect    string
		expectErr string
	}{
		{"12", "@<<<<<", "12    ", ""},
		{"12", "@>>>>>", "    12", ""},
		{"12", "@|||||", "  12  ", ""},
		{"12", "@<<<<:", "12:   ", ""},
		{"12", "@>>>>:", "   12:", ""},
		{"12", "@||||:", " 12:  ", ""},
		{"123456", "@>>>>:", "12345:", ""},
		{"12345678", "@>>>>:", "12345:", ""},
		{"1234\n5678", "@>>>>:", " 1234:", ""},
		{"12", "", "", "invalid format: "},
		{"12", "@xxxxx", "", "invalid string format: @xxxxx"},
	}

	for _, tc := range cases {
		res, err := formatString(tc.value, tc.format)
		if tc.expectErr == "" && err != nil {
			t.Fatalf("Did not expect an error, got: %v", err)
		}
		if tc.expectErr != "" && err.Error() != tc.expectErr {
			t.Fatalf("expected err %s: %v", tc.expectErr, err)
		}
		if res != tc.expect {
			t.Fatalf("expected '%v' got '%v'", tc.expect, res)
		}
	}
}

func TestFormatDataItem(t *testing.T) {
	var cases = []struct {
		name      string
		val       gjson.Result
		format    string
		expect    string
		expectErr string
	}{
		{"null value", gjson.Result{Type: gjson.Null}, "@#####", "?     ", ""},
		{"non existent value", gjson.Result{}, "@###", "?   ", ""},
		{"string value", gjson.Result{Type: gjson.String, Str: "hello"}, "@<<<<<<<<", "hello    ", ""},
		{"true value", gjson.Result{Type: gjson.True}, "@<<<<<<<<", "true     ", ""},
		{"false value", gjson.Result{Type: gjson.False}, "@<<<<<<<<<", "false     ", ""},
		{"number as int", gjson.Result{Type: gjson.Number, Num: 42}, "@####", "42   ", ""},
		{"number as float", gjson.Result{Type: gjson.Number, Num: 3.14}, "@##.##", "3.14  ", ""},
		{"json type errors", gjson.Result{Type: gjson.JSON, Raw: `{"a":1}`}, "@####", "", "invalid type JSON"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := formatDataItem(tc.val, tc.format)
			if tc.expectErr == "" && err != nil {
				t.Fatalf("Did not expect an error, got: %v", err)
			}
			if tc.expectErr != "" && (err == nil || err.Error() != tc.expectErr) {
				t.Fatalf("expected err %s: %v", tc.expectErr, err)
			}
			if res != tc.expect {
				t.Fatalf("expected '%v' got '%v'", tc.expect, res)
			}
		})
	}
}

func TestGetDataItem(t *testing.T) {
	f := &formatter{}

	var cases = []struct {
		name      string
		data      json.RawMessage
		item      string
		expectInt int64
		expectStr string
	}{
		{"simple field", json.RawMessage(`{"row":{"name":"test"}}`), "row.name", 0, "test"},
		{"missing summary defaults to zero", json.RawMessage(`{"report":{"summary":{}}}`), "report.summary.missing", 0, ""},
		{"existing summary value", json.RawMessage(`{"report":{"summary":{"count":5}}}`), "report.summary.count", 5, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := f.getDataItem(tc.data, tc.item)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.expectStr != "" && val.String() != tc.expectStr {
				t.Fatalf("expected '%v' got '%v'", tc.expectStr, val.String())
			}
			if tc.expectStr == "" && val.Int() != tc.expectInt {
				t.Fatalf("expected %d got %d", tc.expectInt, val.Int())
			}
		})
	}
}

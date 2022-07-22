// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"testing"
)

func TestFormatNumber(t *testing.T) {
	var cases = []struct {
		value     interface{}
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
		// formatCommaNumber
		{123456.12, "@,########", "123,456.12", ""},
		// formatBytesNumber
		{123456, "@B#####", "121 KiB", ""},
		// misc
		{123456.123, "@,########", "##########", ""},
		{1.1, "@0.##", "01.10", ""},
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

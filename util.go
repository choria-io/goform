// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package goform

import (
	"bufio"
	"strings"
)

func toLines(body string) ([]string, error) {
	var out []string

	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}
	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return out, nil
}

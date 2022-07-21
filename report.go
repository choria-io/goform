// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type Report struct {
	Name         string `json:"name"`
	BodyLayout   string `json:"body"`
	HeaderLayout string `json:"header"`
	FooterLayout string `json:"footer"`
	RowsPerPage  int    `json:"rows_per_page"`
	bodyLines    []string
	headerLines  []string
	footerLines  []string
	sections     [][]string
	accumulator  map[string]float64
}

type formatter struct {
	bodyLines []string
	acc       map[string]float64
}

func NewFromFile(file string, name string) (*Report, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	r := &Report{}

	err = yaml.Unmarshal(f, r)
	if err != nil {
		return nil, err
	}

	if name != "" {
		r.Name = name
	}
	if r.Name == "" {
		r.Name = "Report"
	}

	if r.RowsPerPage == 0 {
		r.RowsPerPage = 20
	}

	return New(r.Name, r.HeaderLayout, r.BodyLayout, r.FooterLayout, r.RowsPerPage)
}

func New(name string, header string, body string, footer string, rowsPerPage int) (*Report, error) {
	f := &Report{
		Name:         name,
		HeaderLayout: header,
		BodyLayout:   body,
		FooterLayout: footer,
		RowsPerPage:  rowsPerPage,
		accumulator:  map[string]float64{},
	}

	var err error

	f.bodyLines, err = toLines(f.BodyLayout)
	if err != nil {
		return nil, err
	}

	f.headerLines, err = toLines(f.HeaderLayout)
	if err != nil {
		return nil, err
	}

	f.footerLines, err = toLines(f.FooterLayout)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (r *Report) newState(page int, data any, row any, curRow int) map[string]any {
	return map[string]any{
		"report": map[string]any{
			"page":        page,
			"name":        r.Name,
			"current_row": curRow,
			"summary":     r.accumulator,
		},
		"row":  row,
		"data": data,
	}
}

// WriteReportContainedRows retrieves rows from data using query and renders the report, report written to w
func (r *Report) WriteReportContainedRows(w io.Writer, data any, query string) error {
	dj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	val := gjson.GetBytes(dj, query)
	if !val.Exists() {
		return fmt.Errorf("could not find rows at %s", query)
	}

	if !val.IsArray() {
		return fmt.Errorf("%s did not yield an array result", query)
	}

	return r.process(w, data, val.Value().([]any))
}

// ReportContainedRows retrieves rows from data using query and renders the report, returns report as bytes
func (r *Report) ReportContainedRows(data any, query string) ([]byte, error) {
	report := bytes.NewBuffer([]byte{})
	err := r.WriteReportContainedRows(report, data, query)
	if err != nil {
		return nil, err
	}

	return report.Bytes(), nil
}

// Report produce a report of rows, results returned as bytes
func (r *Report) Report(rows []any) ([]byte, error) {
	report := bytes.NewBuffer([]byte{})

	err := r.process(report, nil, rows)
	if err != nil {
		return nil, err
	}

	return report.Bytes(), nil
}

// WriteReport produce a report of rows, result written to w
func (r *Report) WriteReport(w io.Writer, data []any) error {
	return r.process(w, nil, data)
}

func (r *Report) process(w io.Writer, data any, rows []any) error {
	page := 0

	for item, row := range rows {
		if r.HeaderLayout != "" && (page == 0 || (r.RowsPerPage > 0 && item%r.RowsPerPage == 0)) {
			if r.FooterLayout != "" && page != 0 {
				f := &formatter{bodyLines: r.footerLines}
				err := f.process(w, r.newState(page, data, row, item))
				if err != nil {
					return fmt.Errorf("footer processing error: %v", err)
				}
			}

			page++
			f := &formatter{bodyLines: r.headerLines}
			err := f.process(w, r.newState(page, data, row, item))
			if err != nil {
				return fmt.Errorf("header processing error: %v", err)
			}
		}

		f := &formatter{
			bodyLines: r.bodyLines,
			acc:       r.accumulator,
		}

		out, err := f.processBytes(r.newState(page, data, row, item))
		if err != nil {
			return err
		}

		if len(out) > 0 {
			fmt.Fprint(w, string(out))
		}
	}

	if r.FooterLayout != "" {
		f := &formatter{bodyLines: r.footerLines}
		err := f.process(w, r.newState(page, data, nil, len(rows)))
		if err != nil {
			return fmt.Errorf("footer processing error: %v", err)
		}
	}

	return nil
}

func (f *formatter) isPictureLine(line string) bool {
	return strings.ContainsAny(line, "@^")
}

func (f *formatter) countPictureFormats(line string) int {
	return strings.Count(line, "^") + strings.Count(line, "@")
}

func (f *formatter) picturesFromLine(line string) ([]string, error) {
	picts := []string{}

	var inPicture bool
	var curPicture string
	ll := len(line)

	for i, c := range line {
		if c == '@' {
			inPicture = true
		}
		if !inPicture {
			continue
		}

		char := string(c)
		picChar := strings.ContainsAny(char, "@^<|>#.0,B:")

		switch {
		case i == ll-1:
			inPicture = false
			if picChar {
				curPicture += string(c)
			}
			picts = append(picts, curPicture)
			curPicture = ""
		case !picChar:
			inPicture = false
			picts = append(picts, curPicture)
			curPicture = ""
		default:
			curPicture += string(c)
		}
	}

	return picts, nil
}

func (f *formatter) isCommentLine(line string) bool {
	return strings.HasPrefix(line, "#")
}

func (f *formatter) isVariableLine(line string, numVars int) bool {
	return len(strings.Split(line, ",")) == numVars
}

func (f *formatter) variablesFromLine(line string, numVars int) ([]string, error) {
	parts := strings.Split(line, ",")
	if len(parts) != numVars {
		return nil, fmt.Errorf("invalid variable count")
	}

	for i, _ := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts, nil
}

func (f *formatter) accumulate(item string, val float64) {
	name := strings.TrimPrefix(item, "row.")
	if f.acc == nil {
		f.acc = map[string]float64{}
	}

	_, ok := f.acc[name]
	if !ok {
		f.acc[name] = 0
	}

	f.acc[name] += val
}

func (f *formatter) process(w io.Writer, row any) error {
	var (
		numVars   int
		pictures  []string
		variables []string
		fmtLine   string
		err       error
		dj        []byte
	)

	dj, err = json.Marshal(row)
	if err != nil {
		return err
	}

	for i := 0; i < len(f.bodyLines); i++ {
		switch {
		case f.isCommentLine(f.bodyLines[i]):
			continue

		case numVars == 0 && f.isPictureLine(f.bodyLines[i]):
			pictures, err = f.picturesFromLine(f.bodyLines[i])
			if err != nil {
				return fmt.Errorf("invalid format line: %v", err)
			}

			fmtLine = f.bodyLines[i]
			numVars = len(pictures)

		case numVars > 0 && f.isVariableLine(f.bodyLines[i], numVars):
			variables, err = f.variablesFromLine(f.bodyLines[i], numVars)
			if err != nil {
				return fmt.Errorf("invalid variable line: %v", err)
			}

			for i := range pictures {
				val, err := f.getDataItem(dj, variables[i])
				if err != nil {
					return err
				}

				if val.Type == gjson.Number && strings.HasPrefix(variables[i], "row.") {
					f.accumulate(variables[i], val.Float())
				}

				rendered, err := formatDataItem(val, pictures[i])
				if err != nil {
					return fmt.Errorf("could not render item %s into %s: %v", variables[i], pictures[i], err)
				}
				fmtLine = strings.Replace(fmtLine, pictures[i], rendered, 1)
			}

			fmt.Fprintln(w, fmtLine)

			numVars = 0
			variables = nil
			pictures = nil
			fmtLine = ""

		default:
			fmt.Fprintln(w, f.bodyLines[i])
		}
	}

	return nil
}

func (f *formatter) processBytes(row any) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})

	err := f.process(out, row)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

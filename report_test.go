// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package goform

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReport(t *testing.T) {
	header := `
                                                   Sample Report                                                                 Page @##
report.page
---------------------------------------------------------------------------------------------------------------------------------------`
	report := `@##: !@#####! !@<<<<<<<! !@||||||||||||||||! !@>>>>>>! !@#.###! !@0######! !@########! !@##.##! !@0#######! |@B######! !@,#####################!
row.n,row.acc,row.x,row.y,row.z,row.float,row.float,row.float,row.int,row.int,row.gb,row.gb`

	footer := `
@####
report.summary.acc
`

	expected := `
                                                   Sample Report                                                                 Page 1  
---------------------------------------------------------------------------------------------------------------------------------------
0    !1     ! !X thing ! !     Y Thing     ! !Z thing! !3.143 ! !00000003! !3        ! !10.00 ! !000000010! |10 GiB  ! !10,737,430,484         !
1    !1     ! !X thing ! !     Y Thing     ! !Z thing! !3.143 ! !00000003! !3        ! !10.00 ! !000000010! |10 GiB  ! !10,737,430,484         !

2    

                                                   Sample Report                                                                 Page 2  
---------------------------------------------------------------------------------------------------------------------------------------
2    !1     ! !X thing ! !     Y Thing     ! !Z thing! !3.143 ! !00000003! !3        ! !10.00 ! !000000010! |10 GiB  ! !10,737,430,484         !
3    !1     ! !X thing ! !     Y Thing     ! !Z thing! !3.143 ! !00000003! !3        ! !10.00 ! !000000010! |10 GiB  ! !10,737,430,484         !

4    
`

	f, err := New("test", header, report, footer, 2)
	if err != nil {
		t.Fatalf("new failed")
	}

	var data []any
	for i := 0; i < 4; i++ {
		data = append(data, map[string]any{"acc": 1, "n": i, "x": "X thing", "y": "Y Thing", "z": "Z thing", "gb": 10*1024*1024*1024 + 12244, "int": 10, "float": float64(22) / 7})
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if string(out) != expected {
		fmt.Printf("Expected:`%s`\n", expected)
		fmt.Printf("Output:`%s`\n", out)
		t.Fatalf("Did not receive expected outcome")
	}
}

func TestWriteReport(t *testing.T) {
	body := `@<<<<<<<<< @##
row.name, row.age`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	var buf bytes.Buffer
	data := []any{
		map[string]any{"name": "Alice", "age": 30},
		map[string]any{"name": "Bob", "age": 25},
	}

	err = f.WriteReport(&buf, data)
	if err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Fatalf("expected output to contain Alice and Bob, got: %s", out)
	}
}

func TestReportContainedRows(t *testing.T) {
	body := `@<<<<<<<<< @##
row.name, row.age`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	data := map[string]any{
		"people": []any{
			map[string]any{"name": "Alice", "age": 30},
			map[string]any{"name": "Bob", "age": 25},
		},
	}

	out, err := f.ReportContainedRows(data, "people")
	if err != nil {
		t.Fatalf("ReportContainedRows failed: %v", err)
	}

	if !strings.Contains(string(out), "Alice") || !strings.Contains(string(out), "Bob") {
		t.Fatalf("expected output to contain Alice and Bob, got: %s", out)
	}
}

func TestWriteReportContainedRows(t *testing.T) {
	body := `@<<<<<<<<< @##
row.name, row.age`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	t.Run("valid query", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]any{
			"people": []any{
				map[string]any{"name": "Alice", "age": 30},
			},
		}

		err := f.WriteReportContainedRows(&buf, data, "people")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "Alice") {
			t.Fatalf("expected output to contain Alice, got: %s", buf.String())
		}
	})

	t.Run("missing query path", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]any{"other": 1}

		err := f.WriteReportContainedRows(&buf, data, "missing")
		if err == nil || !strings.Contains(err.Error(), "could not find rows") {
			t.Fatalf("expected 'could not find rows' error, got: %v", err)
		}
	})

	t.Run("non array result", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]any{"name": "hello"}

		err := f.WriteReportContainedRows(&buf, data, "name")
		if err == nil || !strings.Contains(err.Error(), "did not yield an array result") {
			t.Fatalf("expected 'did not yield an array result' error, got: %v", err)
		}
	})
}

func TestNewFromFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		td := t.TempDir()
		fp := filepath.Join(td, "report.yaml")
		err := os.WriteFile(fp, []byte(`
name: "My Report"
body: |
  @<<<<<
  row.x
rows_per_page: 10
`), 0644)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		r, err := NewFromFile(fp, "")
		if err != nil {
			t.Fatalf("NewFromFile failed: %v", err)
		}
		if r.Name != "My Report" {
			t.Fatalf("expected name 'My Report' got '%s'", r.Name)
		}
		if r.RowsPerPage != 10 {
			t.Fatalf("expected rows_per_page 10 got %d", r.RowsPerPage)
		}
	})

	t.Run("name override", func(t *testing.T) {
		td := t.TempDir()
		fp := filepath.Join(td, "report.yaml")
		err := os.WriteFile(fp, []byte(`
name: "Original"
body: |
  @<<<<<
  row.x
`), 0644)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		r, err := NewFromFile(fp, "Override")
		if err != nil {
			t.Fatalf("NewFromFile failed: %v", err)
		}
		if r.Name != "Override" {
			t.Fatalf("expected name 'Override' got '%s'", r.Name)
		}
	})

	t.Run("default name", func(t *testing.T) {
		td := t.TempDir()
		fp := filepath.Join(td, "report.yaml")
		err := os.WriteFile(fp, []byte(`
body: |
  @<<<<<
  row.x
`), 0644)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		r, err := NewFromFile(fp, "")
		if err != nil {
			t.Fatalf("NewFromFile failed: %v", err)
		}
		if r.Name != "Report" {
			t.Fatalf("expected default name 'Report' got '%s'", r.Name)
		}
	})

	t.Run("default rows per page", func(t *testing.T) {
		td := t.TempDir()
		fp := filepath.Join(td, "report.yaml")
		err := os.WriteFile(fp, []byte(`
name: test
body: |
  @<<<<<
  row.x
`), 0644)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		r, err := NewFromFile(fp, "")
		if err != nil {
			t.Fatalf("NewFromFile failed: %v", err)
		}
		if r.RowsPerPage != 20 {
			t.Fatalf("expected default rows_per_page 20 got %d", r.RowsPerPage)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewFromFile("/nonexistent/path.yaml", "")
		if err == nil {
			t.Fatalf("expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		td := t.TempDir()
		fp := filepath.Join(td, "bad.yaml")
		err := os.WriteFile(fp, []byte(`{{{not yaml`), 0644)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}

		_, err = NewFromFile(fp, "")
		if err == nil {
			t.Fatalf("expected error for invalid yaml")
		}
	})
}

func TestReportWithCommentLines(t *testing.T) {
	body := `# this is a comment
@<<<<<<<<< @##
row.name, row.age
# another comment`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	data := []any{
		map[string]any{"name": "Alice", "age": 30},
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	if strings.Contains(string(out), "# this is a comment") {
		t.Fatalf("comment lines should not appear in output")
	}
	if !strings.Contains(string(out), "Alice") {
		t.Fatalf("expected output to contain Alice, got: %s", out)
	}
}

func TestReportNoHeader(t *testing.T) {
	body := `@<<<<<<<<< @##
row.name, row.age`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	data := []any{
		map[string]any{"name": "Alice", "age": 30},
		map[string]any{"name": "Bob", "age": 25},
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	if strings.Contains(string(out), "Page") {
		t.Fatalf("should not contain header, got: %s", out)
	}
}

func TestReportNoPagination(t *testing.T) {
	header := `--- Page @## ---
report.page`
	body := `@<<<<<<<<<<
row.name`

	f, err := New("test", header, body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	var data []any
	for i := 0; i < 5; i++ {
		data = append(data, map[string]any{"name": fmt.Sprintf("row%d", i)})
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	if strings.Count(string(out), "--- Page") != 1 {
		t.Fatalf("expected exactly 1 header with RowsPerPage=0, got: %s", out)
	}
}

func TestReportMissingField(t *testing.T) {
	body := `@<<<<<<<<< @######
row.name, row.missing`

	f, err := New("test", "", body, "", 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	data := []any{
		map[string]any{"name": "Alice"},
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	if !strings.Contains(string(out), "?") {
		t.Fatalf("expected '?' for missing field, got: %s", out)
	}
}

func TestReportAccumulator(t *testing.T) {
	body := `@<<<<<<<<< @##
row.name, row.score`
	footer := `Total: @####
report.summary.score`

	f, err := New("test", "", body, footer, 0)
	if err != nil {
		t.Fatalf("new failed: %v", err)
	}

	data := []any{
		map[string]any{"name": "Alice", "score": 10},
		map[string]any{"name": "Bob", "score": 20},
	}

	out, err := f.Report(data)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	if !strings.Contains(string(out), "30") {
		t.Fatalf("expected accumulated score of 30 in footer, got: %s", out)
	}
}

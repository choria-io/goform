// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
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

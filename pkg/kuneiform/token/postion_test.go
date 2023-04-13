package token

import (
	"reflect"
	"testing"
)

func TestFile_Position(t *testing.T) {
	length := 75
	lines := []int{0, 10, 23, 31, 46, 53, 66, 70}

	file := &File{
		//Name:  tt.fields.Name,
		Size:  length,
		Lines: lines,
	}
	tests := []struct {
		name string
		args Pos
		want Position
	}{
		{"start of the file", 0, Position{Pos(1), Pos(1)}},
		{"begin of a line", 66, Position{Pos(7), Pos(1)}},
		{"middle of a line", 68, Position{Pos(7), Pos(3)}},
		{"end of a line", 69, Position{Pos(7), Pos(4)}},
		{"end of the file", 75, Position{Pos(8), Pos(6)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := file.Position(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Position() = %v, want %v", got, tt.want)
			}
		})
	}
}

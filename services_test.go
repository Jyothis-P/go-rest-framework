package grf

import "testing"

func TestGetPluralTableDriven(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"pointer plural test", "*models.Todo", "todos"},
		{"pointer slice plural test", "*[]models.Todo", "todos"},
		{"base plural test", "models.Todo", "todos"},
		{"es plural test", "*models.Box", "boxes"},
		{"package plural test", "main.Box", "boxes"},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := getPlural(tt.input)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

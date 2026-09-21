package tcp

import "testing"

func TestAveragePopCount(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float32
	}{
		{name: "empty", want: 0},
		{name: "zero", data: []byte{0, 0}, want: 0},
		{name: "full", data: []byte{0xff}, want: 8},
		{name: "mixed", data: []byte{0x00, 0xff}, want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := averagePopCount(tt.data); got != tt.want {
				t.Fatalf("averagePopCount(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestContiguousPrintable(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{name: "empty"},
		{name: "middle", data: []byte{0, 'a', 'b', 0, 'c'}, want: 2},
		{name: "end", data: []byte{'a', 0, 'b', 'c', 'd'}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contiguousPrintable(tt.data); got != tt.want {
				t.Fatalf("contiguousPrintable(%v) = %d, want %d", tt.data, got, tt.want)
			}
		})
	}
}

package fare

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		name                    string
		distance                int32
		base, rate, minimum     int64
		multiplier              int32
		wantDistance, wantTotal int64
	}{{"distance fare", 120000, 10000, 300, 25000, 10000, 36000, 46000}, {"minimum fare", 30000, 10000, 300, 25000, 10000, 9000, 25000}, {"class multiplier", 120000, 10000, 300, 25000, 15000, 36000, 69000}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance, total := Calculate(tt.distance, tt.base, tt.rate, tt.minimum, tt.multiplier)
			if distance != tt.wantDistance || total != tt.wantTotal {
				t.Fatalf("got %d/%d, want %d/%d", distance, total, tt.wantDistance, tt.wantTotal)
			}
		})
	}
}

package datetime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sumup/sumup-go/datetime"
)

func TestDate_String(t *testing.T) {
	tests := []struct {
		name     string
		date     datetime.Date
		expected string
	}{
		{
			name:     "valid date",
			date:     datetime.Date{time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)},
			expected: "2023-10-15",
		},
		{
			name:     "first day of year",
			date:     datetime.Date{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			expected: "2024-01-01",
		},
		{
			name:     "last day of year",
			date:     datetime.Date{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
			expected: "2024-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.String(); got != tt.expected {
				t.Errorf("Date.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		date     datetime.Date
		expected string
	}{
		{
			name:     "valid date",
			date:     datetime.Date{time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)},
			expected: `"2023-10-15"`,
		},
		{
			name:     "leap year date",
			date:     datetime.Date{time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
			expected: `"2024-02-29"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.MarshalJSON()
			if err != nil {
				t.Errorf("Date.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("Date.MarshalJSON() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    datetime.Date
		wantErr bool
	}{
		{
			name:    "valid date",
			input:   `"2023-10-15"`,
			want:    datetime.Date{time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
		},
		{
			name:    "leap year date",
			input:   `"2024-02-29"`,
			want:    datetime.Date{time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
		},
		{
			name:    "invalid date format",
			input:   `"2023/10/15"`,
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   `"2023-13-01"`,
			wantErr: true,
		},
		{
			name:    "not a date",
			input:   `"hello"`,
			wantErr: true,
		},
		{
			name:    "missing quotes",
			input:   `2023-10-15`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got datetime.Date
			err := got.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Date.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want.Time) {
				t.Errorf("Date.UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDate_JSONRoundTrip(t *testing.T) {
	type testStruct struct {
		Date datetime.Date `json:"date"`
	}

	original := testStruct{
		Date: datetime.Date{time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)},
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded testStruct
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !decoded.Date.Equal(original.Date.Time) {
		t.Errorf("Round trip failed: got %v, want %v", decoded.Date, original.Date)
	}
}

func TestTime_String(t *testing.T) {
	tests := []struct {
		name     string
		time     datetime.Time
		expected string
	}{
		{
			name:     "valid time",
			time:     datetime.Time{time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
			expected: "14:30:45",
		},
		{
			name:     "midnight",
			time:     datetime.Time{time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
			expected: "00:00:00",
		},
		{
			name:     "almost midnight",
			time:     datetime.Time{time.Date(0, 1, 1, 23, 59, 59, 0, time.UTC)},
			expected: "23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.time.String(); got != tt.expected {
				t.Errorf("Time.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		time     datetime.Time
		expected string
	}{
		{
			name:     "valid time",
			time:     datetime.Time{time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
			expected: `"14:30:45"`,
		},
		{
			name:     "midnight",
			time:     datetime.Time{time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
			expected: `"00:00:00"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.time.MarshalJSON()
			if err != nil {
				t.Errorf("Time.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("Time.MarshalJSON() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    datetime.Time
		wantErr bool
	}{
		{
			name:    "valid time",
			input:   `"14:30:45"`,
			want:    datetime.Time{time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
			wantErr: false,
		},
		{
			name:    "midnight",
			input:   `"00:00:00"`,
			want:    datetime.Time{time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
		},
		{
			name:    "almost midnight",
			input:   `"23:59:59"`,
			want:    datetime.Time{time.Date(0, 1, 1, 23, 59, 59, 0, time.UTC)},
			wantErr: false,
		},
		{
			name:    "invalid time format",
			input:   `"14-30-45"`,
			wantErr: true,
		},
		{
			name:    "invalid hour",
			input:   `"25:00:00"`,
			wantErr: true,
		},
		{
			name:    "invalid minute",
			input:   `"14:60:00"`,
			wantErr: true,
		},
		{
			name:    "invalid second",
			input:   `"14:30:60"`,
			wantErr: true,
		},
		{
			name:    "not a time",
			input:   `"hello"`,
			wantErr: true,
		},
		{
			name:    "missing quotes",
			input:   `14:30:45`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got datetime.Time
			err := got.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Time.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want.Time) {
				t.Errorf("Time.UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_JSONRoundTrip(t *testing.T) {
	type testStruct struct {
		Time datetime.Time `json:"time"`
	}

	original := testStruct{
		Time: datetime.Time{time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded testStruct
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !decoded.Time.Equal(original.Time.Time) {
		t.Errorf("Round trip failed: got %v, want %v", decoded.Time, original.Time)
	}
}

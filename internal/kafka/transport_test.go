package kafka

import "testing"

func TestTLSServerName(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		want    string
	}{
		{
			name:    "host port",
			brokers: []string{"broker.example.com:9092"},
			want:    "broker.example.com",
		},
		{
			name:    "host only",
			brokers: []string{"broker.example.com"},
			want:    "broker.example.com",
		},
		{name: "ipv6 brackets", brokers: []string{"[::1]:9092"}, want: "::1"},
		{name: "empty", brokers: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tlsServerName(tt.brokers); got != tt.want {
				t.Fatalf("tlsServerName() = %q, want %q", got, tt.want)
			}
		})
	}
}

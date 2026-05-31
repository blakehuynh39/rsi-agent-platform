package dbread

import "testing"

func TestDSNFromSecretString(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		want    string
		wantErr bool
	}{
		{
			name:   "plain DSN",
			secret: "postgres://readonly@example/db?sslmode=require",
			want:   "postgres://readonly@example/db?sslmode=require",
		},
		{
			name:   "JSON dsn",
			secret: `{"dsn":"postgres://readonly@example/db?sslmode=require","port":5432}`,
			want:   "postgres://readonly@example/db?sslmode=require",
		},
		{
			name:    "JSON DSN field must be string",
			secret:  `{"dsn":5432,"host":"example"}`,
			wantErr: true,
		},
		{
			name:    "JSON missing DSN field",
			secret:  `{"host":"example","port":5432}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dsnFromSecretString(tt.secret)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got DSN %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("DSN = %q, want %q", got, tt.want)
			}
		})
	}
}

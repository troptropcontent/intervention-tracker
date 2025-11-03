package routers

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
)

type formTestStruct struct {
	Key string `form:"key"`
}

func TestParseFormData(t *testing.T) {
	tests := []struct {
		name    string
		c       echo.Context
		target  any
		want    any
		wantErr bool
	}{
		{
			name:    "regular form data",
			c:       tests.CreateEchoContextWithFormData(map[string]string{"key": "value"}),
			target:  &formTestStruct{},
			want:    &formTestStruct{Key: "value"},
			wantErr: false,
		},
		{
			name:    "multipart form data",
			c:       tests.CreateEchoContextWithMultipartData(map[string]string{"key": "value"}, map[string][]byte{}),
			target:  &formTestStruct{},
			want:    &formTestStruct{Key: "value"},
			wantErr: false,
		},
		{
			name:    "empty form data",
			c:       tests.CreateEchoContextWithFormData(map[string]string{}),
			target:  &formTestStruct{},
			want:    &formTestStruct{},
			wantErr: false,
		},
		{
			name:    "invalid multipart form - malformed body",
			c:       tests.CreateEchoContextWithInvalidMultipartForm(),
			target:  &formTestStruct{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid form data - corrupted body",
			c:       tests.CreateEchoContextWithCorruptedForm(),
			target:  &formTestStruct{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "decode error - wrong target type",
			c:       tests.CreateEchoContextWithFormData(map[string]string{"key": "value"}),
			target:  "not a pointer to struct",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseFormData(tt.c, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, tt.target)
			}
		})
	}
}

package routers

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"gorm.io/gorm"
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

func TestFindAuthenticatedUser(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func(db *gorm.DB) (ctx echo.Context, err error)
		setupDB   func(*gorm.DB) *models.User
		setupCtx  func(echo.Context, *models.User)
		wantErr   bool
		errMsg    string
		checkUser func(*testing.T, *models.User)
	}{
		{
			name: "successfully finds authenticated user",
			setup: func(db *gorm.DB) (ctx echo.Context, err error) {
				user := factories.NewUser().Create(db)
				context := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).Build()
				return context, nil
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
				assert.True(t, user.IsActive)
			},
		},
		{
			name: "returns error when user_id not in context",
			setup: func(db *gorm.DB) (ctx echo.Context, err error) {
				context := tests.NewContext(http.MethodPost, "/").Build()
				return context, nil
			},
			wantErr: true,
			errMsg:  "no authenticated user in context",
		},
		{
			name: "returns error when user not found in database",
			setup: func(db *gorm.DB) (ctx echo.Context, err error) {
				user := factories.NewUser().Build()
				user.ID = 0
				context := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).Build()
				return context, nil
			},
			wantErr: true,
			errMsg:  "could not find authenticated user",
		},
		{
			name: "does not return soft-deleted users",
			setup: func(db *gorm.DB) (ctx echo.Context, err error) {
				user := factories.NewUser().Create(db)
				context := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).Build()
				db.Delete(user)
				return context, nil
			},
			wantErr: true,
			errMsg:  "could not find authenticated user",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			db := tests.SetupTestDB(t)

			// Setup
			ctx, err := tt.setup(db)
			require.NoError(t, err)

			// Execute function
			user, err := FindAuthenticatedUser(ctx, db)

			// Check error expectation
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				if tt.checkUser != nil {
					tt.checkUser(t, user)
				}
			}
		})
	}
}

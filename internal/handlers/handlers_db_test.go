package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

// Database integration tests - require actual database connection
func TestHandlers_GetPortal_WithRealDB(t *testing.T) {
	// Skip if not in integration test mode or no DB available
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Try to connect to test database
	db, err := database.Connect()
	if err != nil {
		t.Skip("Skipping database test - database not available:", err)
	}
	defer db.Close()

	// Create test portal data
	testUUID := "123e4567-e89b-12d3-a456-426614174000"
	testPortal := models.Portal{
		UUID:              testUUID,
		Name:              "Test Portal",
		AddressStreet:     "123 Test Street",
		AddressZipcode:    "75001",
		AddressCity:       "Paris",
		ContractorCompany: "Test Corp",
		ContactPhone:      "0123456789",
		ContactEmail:      "test@example.com",
		InstallationDate:  time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Insert test data
	_, err = db.NamedExec(`
		INSERT INTO portals (uuid, name, address_street, address_zipcode, address_city, 
		                    contractor_company, contact_phone, contact_email, 
		                    installation_date, created_at, updated_at) 
		VALUES (:uuid, :name, :address_street, :address_zipcode, :address_city,
		        :contractor_company, :contact_phone, :contact_email,
		        :installation_date, :created_at, :updated_at)
		ON CONFLICT (uuid) DO UPDATE SET
		name = :name, updated_at = :updated_at
	`, testPortal)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		db.Exec("DELETE FROM portals WHERE uuid = $1", testUUID)
	}()

	t.Run("GetPortal_Success", func(t *testing.T) {
		// Setup
		h := &Handlers{DB: db}
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/portals/"+testUUID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("uuid")
		c.SetParamValues(testUUID)

		// Execute
		err := h.GetPortal(c)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, "Test Portal")
		assert.Contains(t, body, "123 Test Street")
		assert.Contains(t, body, "75001")
		assert.Contains(t, body, "Paris")
	})

	t.Run("GetPortal_NotFound", func(t *testing.T) {
		// Setup
		h := &Handlers{DB: db}
		e := echo.New()
		nonExistentUUID := "00000000-0000-0000-0000-000000000000"
		req := httptest.NewRequest(http.MethodGet, "/portals/"+nonExistentUUID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("uuid")
		c.SetParamValues(nonExistentUUID)

		// Execute
		err := h.GetPortal(c)

		// Assert
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
		assert.Equal(t, "Portal not found", httpErr.Message)
	})
}

func TestHandlers_Database_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database connection test in short mode")
	}

	// Test database connection
	db, err := database.Connect()
	if err != nil {
		t.Skip("Database not available for testing:", err)
	}
	defer db.Close()

	// Test basic query
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM portals")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)

	// Test database schema
	var tableExists bool
	err = db.Get(&tableExists, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'portals'
		)
	`)
	assert.NoError(t, err)
	assert.True(t, tableExists, "portals table should exist")

	// Test required columns exist
	requiredColumns := []string{
		"id", "uuid", "name", "address_street", "address_zipcode",
		"address_city", "contractor_company", "contact_phone",
		"contact_email", "installation_date", "created_at", "updated_at",
	}

	for _, column := range requiredColumns {
		var columnExists bool
		err = db.Get(&columnExists, `
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'portals' 
				AND column_name = $1
			)
		`, column)
		assert.NoError(t, err)
		assert.True(t, columnExists, "Column %s should exist in portals table", column)
	}
}

func TestHandlers_Portal_CRUD_Operations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CRUD test in short mode")
	}

	db, err := database.Connect()
	if err != nil {
		t.Skip("Database not available for testing:", err)
	}
	defer db.Close()

	testUUID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID format

	// Clean up before and after
	cleanup := func() {
		db.Exec("DELETE FROM portals WHERE uuid = $1", testUUID)
	}
	cleanup()
	defer cleanup()

	t.Run("Insert_Portal", func(t *testing.T) {
		portal := models.Portal{
			UUID:              testUUID,
			Name:              "CRUD Test Portal",
			AddressStreet:     "456 CRUD Street",
			AddressZipcode:    "75002",
			AddressCity:       "Paris",
			ContractorCompany: "CRUD Corp",
			ContactPhone:      "0123456789",
			ContactEmail:      "crud@example.com",
			InstallationDate:  time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		_, err := db.NamedExec(`
			INSERT INTO portals (uuid, name, address_street, address_zipcode, address_city,
			                    contractor_company, contact_phone, contact_email,
			                    installation_date, created_at, updated_at)
			VALUES (:uuid, :name, :address_street, :address_zipcode, :address_city,
			        :contractor_company, :contact_phone, :contact_email,
			        :installation_date, :created_at, :updated_at)
		`, portal)

		assert.NoError(t, err)
	})

	t.Run("Read_Portal", func(t *testing.T) {
		var portal models.Portal
		err := db.Get(&portal, "SELECT * FROM portals WHERE uuid = $1", testUUID)
		
		assert.NoError(t, err)
		assert.Equal(t, testUUID, portal.UUID)
		assert.Equal(t, "CRUD Test Portal", portal.Name)
		assert.Equal(t, "456 CRUD Street", portal.AddressStreet)
		assert.Equal(t, "75002", portal.AddressZipcode)
		assert.Equal(t, "Paris", portal.AddressCity)
	})

	t.Run("Handler_GetPortal_CRUD", func(t *testing.T) {
		h := &Handlers{DB: db}
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/portals/"+testUUID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("uuid")
		c.SetParamValues(testUUID)

		err := h.GetPortal(c)
		
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, "CRUD Test Portal")
		assert.Contains(t, body, "456 CRUD Street")
		assert.Contains(t, body, "75002")
	})
}
package devicesCrud

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	connectionString string
}

// for running test needs to start with Test
type ServicesTestSuite struct {
	suite.Suite
	db          *sql.DB
	pgContainer *PostgresContainer
	ctx         context.Context
}

// runs at the beginning before any test have run yet
// sets up the containter and the db in that container
func (suite *ServicesTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// create the actual test container
	pgContainer, err := createPostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	// connect to the test container database
	db, err := sql.Open("postgres", suite.pgContainer.connectionString)
	if err != nil {
		log.Fatal("Could not connect to database")
	}
	suite.db = db
}

// takes a snapshot of the initial state of the database before every test
func (suite *ServicesTestSuite) SetupTest() {
	err := suite.pgContainer.Snapshot(suite.ctx)
	if err != nil {
		err = suite.pgContainer.Terminate(suite.ctx)
		if err != nil {
			log.Fatalf("error terminating postgres container: %s", err)
		}
		log.Fatal(err)
	}
}

// drops the database and restores it to the initial state using a snapshot after every test is done
func (suite *ServicesTestSuite) TearDownTest() {
	err := suite.pgContainer.Restore(suite.ctx)
	if err != nil {
		err = suite.pgContainer.Terminate(suite.ctx)
		if err != nil {
			log.Fatalf("error terminating postgres container: %s", err)
		}
		log.Fatal(err)
	}
}

// runs after all test done
// closes db then tears down container
func (suite *ServicesTestSuite) TearDownSuite() {
	err := suite.db.Close()
	if err != nil {
		log.Fatalf("error terminating closing db connection: %s", err)
	}
	err = suite.pgContainer.Terminate(suite.ctx)
	if err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *ServicesTestSuite) TestLightDeviceAddEmptyDb() {
	light := newLightDevice("unique", "light1", "light",
		"http._tcp", "custom", "setunique",
		"getunique", "unique.local", nil, false, false)

	err := AddLightDevice(suite.db, *light)
	assert.Equal(suite.T(), nil, err)

	numLightDevices, err := getNumberOfItemsFromTable(suite.db, "light")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 1, numLightDevices)

	numDevices, err := getNumberOfItemsFromTable(suite.db, "device")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 1, numDevices)

	fetchedLight, err := getLightDevice(suite.db, "unique")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), true, EqualLightDevices(light, fetchedLight))
}

func (suite *ServicesTestSuite) TestLightDeviceAddDuplicate() {
	light := newLightDevice("unique", "light1", "light",
		"http._tcp", "custom", "setunique",
		"getunique", "unique.local", nil, false, false)

	err := AddLightDevice(suite.db, *light)
	assert.Equal(suite.T(), nil, err)

	duplicateLight := newLightDevice("unique", "light1", "light",
		"http._tcp", "custom", "setunique",
		"getunique", "unique.local", nil, false, false)
	err = AddLightDevice(suite.db, *duplicateLight)

	assert.NotEqual(suite.T(), nil, err)
	var notUniqueError ErrorDuplicateData
	assert.ErrorAs(suite.T(), err, &notUniqueError)
	numLightDevices, err := getNumberOfItemsFromTable(suite.db, "light")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 1, numLightDevices)

	numDevices, err := getNumberOfItemsFromTable(suite.db, "device")
	assert.Equal(suite.T(), nil, err)
	assert.Equal(suite.T(), 1, numDevices)
}

func (suite *ServicesTestSuite) TestLightDeviceAddNonValidNull() {
	type testCase struct {
		name         string
		nullifyField func(*LightDevice)
	}

	testCases := []testCase{
		{"null id", func(l *LightDevice) { l.DeviceID = nil }},
		{"null name", func(l *LightDevice) { l.DeviceName = nil }},
		{"null type", func(l *LightDevice) { l.DeviceType = nil }},
		{"null servicetype", func(l *LightDevice) { l.ServiceType = nil }},
		{"null manufactor", func(l *LightDevice) { l.Manufactor = nil }},
		{"null settopic", func(l *LightDevice) { l.SetTopic = nil }},
		{"null gettopic", func(l *LightDevice) { l.GetTopic = nil }},
		{"null isdimmable", func(l *LightDevice) { l.IsDimmable = nil }},
		{"null isrgb", func(l *LightDevice) { l.IsRgb = nil }},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Setup
			light := newLightDevice("unique", "light1", "light",
				"http._tcp", "custom", "setunique",
				"getunique", "unique.local", nil, false, false)

			tc.nullifyField(light)

			// Execute
			err := AddLightDevice(suite.db, *light)

			// Assert error is a not-null violation
			var nullNotAllowedError ErrorNotNullViolation
			assert.ErrorAs(t, err, &nullNotAllowedError)

			// Assert nothing was inserted
			numLights, err := getNumberOfItemsFromTable(suite.db, "light")
			assert.NoError(t, err)
			assert.Equal(t, 0, numLights)
		})
	}
}

func (suite *ServicesTestSuite) TestLightDeviceAddEmptyStrings() {
	type testCase struct {
		name               string
		emptifyStringField func(*LightDevice)
	}

	testCases := []testCase{
		{"empty id", func(l *LightDevice) { id := ""; l.DeviceID = &id }},
		{"empty name", func(l *LightDevice) { name := ""; l.DeviceName = &name }},
		{"empty settopic", func(l *LightDevice) { setTopic := ""; l.SetTopic = &setTopic }},
		{"empty gettopic", func(l *LightDevice) { getTopic := ""; l.GetTopic = &getTopic }},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Setup
			light := newLightDevice("unique", "light1", "light",
				"http._tcp", "custom", "setunique",
				"getunique", "unique.local", nil, false, false)

			tc.emptifyStringField(light)

			// Execute
			err := AddLightDevice(suite.db, *light)
			fmt.Print(light.DeviceID)

			// Assert error is a not-null violation
			var valueNotAllowedError ErrorIllegalData
			assert.ErrorAs(t, err, &valueNotAllowedError)

			// Assert nothing was inserted
			numLights, err := getNumberOfItemsFromTable(suite.db, "light")
			assert.NoError(t, err)
			assert.Equal(t, 0, numLights)
		})
	}
}

func (suite *ServicesTestSuite) TestFetchAllLights() {
	light1 := newLightDevice("light1", "light1", "light",
		"http._tcp", "custom", "set1", "get1", "light1.local", nil, false, false)
	light2 := newLightDevice("light2", "light2", "light",
		"http._tcp", "custom", "set2", "get2", "light2.local", nil, false, false)
	err := AddLightDevice(suite.db, *light1)
	assert.NoError(suite.T(), err)
	err = AddLightDevice(suite.db, *light2)
	assert.NoError(suite.T(), err)
	lights, err := GetAllLightDevices(suite.db)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(lights))
}

// This is what runs the actual test in the suite
func TestServicesTestSuite(t *testing.T) {
	suite.Run(t, new(ServicesTestSuite))
}

func createPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	postgresContainer, err := postgres.Run(ctx, "postgres:14.8-alpine",
		postgres.WithInitScripts(filepath.Join(".", "init-db.sql")),
		postgres.WithDatabase("smarthome"),
		postgres.WithUsername("emmanuelbastidas"),
		postgres.WithPassword("marcos"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		return nil, err
	}
	pgContainer := PostgresContainer{PostgresContainer: postgresContainer, connectionString: connStr}
	err = pgContainer.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &pgContainer, nil
}

func getNumberOfItemsFromTable(db *sql.DB, tableName string) (int, error) {
	// NOTE THIS IS NOT SAFE BUT IT IS THE ONLY WAY TO DO THIS
	countAllDevicesStmt := "SELECT count(*) FROM " + tableName
	res, err := db.Query(countAllDevicesStmt)
	if err != nil {
		return -1, err
	}

	defer res.Close()
	var count int

	res.Next()
	err = res.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func getLightDevice(db *sql.DB, id string) (*LightDevice, error) {
	query := "SELECT device.id, name, servicetype, devicetype, manufactor, settopic, gettopic, endpoint, room, dimmable, rgb FROM light join device on device.id = light.id where device.id = $1"

	var (
		deviceID    string
		deviceName  string
		serviceType string
		deviceType  string
		manufactor  string
		setTopic    string
		getTopic    string
		endPoint    string
		roomID      sql.NullInt64
		isDimmable  bool
		isRgb       bool
	)

	err := db.QueryRow(query, id).Scan(
		&deviceID, &deviceName, &serviceType, &deviceType,
		&manufactor, &setTopic, &getTopic, &endPoint,
		&roomID, &isDimmable, &isRgb,
	)
	if err != nil {
		return nil, err
	}

	light := LightDevice{
		DeviceID:    &deviceID,
		DeviceName:  &deviceName,
		ServiceType: &serviceType,
		DeviceType:  &deviceType,
		Manufactor:  &manufactor,
		SetTopic:    &setTopic,
		GetTopic:    &getTopic,
		EndPoint:    &endPoint,
		IsDimmable:  &isDimmable,
		IsRgb:       &isRgb,
	}

	if roomID.Valid {
		roomVal := int(roomID.Int64)
		light.RoomID = &roomVal
	}

	return &light, nil
}

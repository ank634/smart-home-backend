package devicesCrud

//all functions that are used for handling http requests relation to devices crud
import (
	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// AddDevice adds a iot device to our home automation system
// todo add mdns device check maybe a ping
// todo maybe pass values or interface instead of struct
func AddDevice(db *sql.DB, device SmartHomeDevice) error {
	query := "INSERT INTO device VALUES ($1, $2, $3, $4, $5, $6, $7)"

	// this executes the db query should not really be used with select
	// use only if the database operation return no operation
	// TODO: decide if I should check for existing value and how to get this to the user
	// or leave that as an error
	_, err := db.Exec(query, device.DeviceID, device.DeviceName,
		device.DeviceType, device.ServiceType,
		device.SetTopic, device.GetTopic,
		device.DeviceUrl)

	return err
}

// todo add mdns device check maybe a ping
// todo maybe pass values or interface instead of struct
func DeleteDevice(db *sql.DB, device SmartHomeDevicePatch) (bool, error) {
	query := "DELETE FROM device WHERE DEVICEID = $1"
	// TODO make this a transaction so all has to succeed-
	result, err := db.Exec(query, device.DeviceID)
	if err != nil {
		return false, err
	}

	var rowsAffected int64
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// EditDevice will attempt to edit the name of device. Will return true if sucessfuly updated
// false if it does not exist in order to facilitate 404
func EditDevice(db *sql.DB, deviceId string, device SmartHomeDevicePatch) (bool, error) {
	query := "UPDATE device SET devicename = $1 WHERE deviceid = $2"
	result, err := db.Exec(query, device.DeviceName, deviceId)
	if err != nil {
		return false, err
	}

	var rowsAffected int64
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func GetDevices(db *sql.DB) ([]SmartHomeDevice, error) {
	query := "SELECT * FROM device"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	// if you only do var devices []SmartHomeDevice it init to nil
	// this init to empty array
	var devices []SmartHomeDevice = []SmartHomeDevice{}
	defer rows.Close()
	for rows.Next() {
		var tempDevice SmartHomeDevice
		err = rows.Scan(&tempDevice.DeviceID, &tempDevice.DeviceName,
			&tempDevice.DeviceType, &tempDevice.ServiceType,
			&tempDevice.SetTopic, &tempDevice.GetTopic,
			&tempDevice.DeviceUrl)

		if err != nil {
			return nil, err
		}
		devices = append(devices, tempDevice)
	}
	return devices, nil
}

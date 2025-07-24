package devicesCrud

//all functions that are used for handling http requests relation to devices crud
import (
	"database/sql"

	"github.com/lib/pq"
)

// ///// LIGHT //////////////
func AddLightDevice(db *sql.DB, light LightDevice) error {
	insertionDeviceTableStatement := "INSERT INTO device(id, name, servicetype, devicetype, manufactor, settopic, gettopic, endpoint, room) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(insertionDeviceTableStatement, light.DeviceID, light.DeviceName,
		light.ServiceType, light.DeviceType, light.Manufactor,
		light.SetTopic, light.GetTopic, light.EndPoint,
		light.RoomID)

	// refactor me
	if err != nil {
		tx.Rollback()
		err, ok := err.(*pq.Error)
		if !ok {
			return err
		}
		if err.Code == "23502" {
			return ErrorNotNullViolation{"This value may not be null"}
		}

		if err.Code == "23505" {
			return ErrorDuplicateData{"This value is not unique"}
		}
		if err.Code == "23514" || err.Code == "22P02" || err.Code == "23503" {
			return ErrorIllegalData{err.Error()}
		}
		return err
	}

	insertLightTableStatement := "Insert into light(id, dimmable, rgb) VALUES($1, $2, $3)"

	_, err = tx.Exec(insertLightTableStatement, light.DeviceID, light.IsDimmable, light.IsRgb)
	if err != nil {
		tx.Rollback()
		err, ok := err.(*pq.Error)
		if !ok {
			return err
		}
		if err.Code == "23502" {
			return ErrorNotNullViolation{"This value may not be null"}
		}

		if err.Code == "23505" {
			return ErrorDuplicateData{"This value is not unique"}
		}
		if err.Code == "23514" {
			return ErrorIllegalData{"Data value not allowed"}
		}
		return err
	}

	err = tx.Commit()
	return err
}

func GetAllLightDevices(db *sql.DB) ([]LightDevice, error) {
	query := `SELECT device.id, name, servicetype, devicetype,
		manufactor, settopic, gettopic, endpoint, room, dimmable, rgb 
		FROM DEVICE JOIN LIGHT 
		ON device.id = light.id`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lights []LightDevice

	for rows.Next() {
		var light LightDevice
		var roomID sql.NullInt64

		err := rows.Scan(
			&light.DeviceID, &light.DeviceName, &light.ServiceType,
			&light.DeviceType, &light.Manufactor, &light.SetTopic,
			&light.GetTopic, &light.EndPoint, &roomID,
			&light.IsDimmable, &light.IsRgb,
		)
		if err != nil {
			return nil, err
		}

		if roomID.Valid {
			roomVal := int(roomID.Int64)
			light.RoomID = &roomVal
		}

		lights = append(lights, light)
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lights, nil
}

// /// GENERIC ////////////////
// todo add mdns device check maybe a ping
// todo maybe pass values or interface instead of struct
func DeleteDevice(db *sql.DB, id string) (bool, error) {
	query := "DELETE FROM device WHERE id = $1"
	// TODO make this a transaction so all has to succeed-
	result, err := db.Exec(query, id)
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
	query := "UPDATE device SET name = $1 WHERE id = $2"
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

func GetAllDevices(db *sql.DB) ([]SmartHomeDevice, error) {
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
			&tempDevice.EndPoint)

		if err != nil {
			return nil, err
		}
		devices = append(devices, tempDevice)
	}
	return devices, nil
}

func GetDevicesByServiceType(db *sql.DB, serviceType string) ([]SmartHomeDevice, error) {
	query := "SELECT * FROM device WHERE servicetype = $1"
	rows, err := db.Query(query, serviceType)
	if err != nil {
		return nil, err
	}

	var devices []SmartHomeDevice = []SmartHomeDevice{}

	defer rows.Close()
	for rows.Next() {
		var tempDevice SmartHomeDevice
		err = rows.Scan(&tempDevice.DeviceID, &tempDevice.DeviceName,
			&tempDevice.DeviceType, &tempDevice.ServiceType,
			&tempDevice.SetTopic, &tempDevice.GetTopic,
			&tempDevice.EndPoint)

		if err != nil {
			return nil, err
		}
		devices = append(devices, tempDevice)
	}
	return devices, nil
}

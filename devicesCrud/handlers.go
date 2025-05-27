package devicesCrud

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
)

// AddDeviceHandler this lets it have dependency injection and inject
func AddDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		// by default the decoder will not throw an error even if the wrong params are passed or no params
		decoder := json.NewDecoder(req.Body)
		// this is what makes it throw an error if bad params
		decoder.DisallowUnknownFields()

		var newDevice SmartHomeDevice
		// when decoding any values not present will just be set to the default empty values
		err := decoder.Decode(&newDevice)

		if err != nil {
			http.Error(w, "Improper parameters", 400)
			return
		}
		// if any is a default value then check and make sure
		if newDevice.DeviceID == "" ||
			newDevice.DeviceName == "" ||
			newDevice.DeviceType == "" ||
			newDevice.DeviceUrl == "" ||
			newDevice.GetTopic == "" ||
			newDevice.SetTopic == "" {
			http.Error(w, "error: missing parameters", 400)
			return
		}

		err = AddDevice(db, newDevice)

		// type assertions allow us to access the concrete type if we only have access to its
		// interface error is an interface and we have that for the db driver what implements
		// the interface is the concrete type *pq.Error so for better error messages we can
		// access the concrete type
		// the package is pq and what we want to access is there Error struct that implements err
		if err != nil {
			pqErr, ok := err.(*pq.Error)
			// if type assertion worked out
			if ok {
				// meant for when not accurate
				//https://github.com/lib/pq/blob/b7ffbd3b47da4290a4af2ccd253c74c2c22bfabf/error.go#L11
				if pqErr.Code == "23505" {
					http.Error(w, "Must be unique iot device", 400)
					return
				}
			}
			http.Error(w, "internal server error", 500)
			return
		}

		// returns ok if everything good
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("You Added a device"))
	}
}

func EditDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		deviceId := req.PathValue("id")
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()

		var newDevice SmartHomeDevicePatch
		err := decoder.Decode(&newDevice)
		if err != nil {
			http.Error(w, "Improper parameters", 400)
			return
		}

		if deviceId == "" || newDevice.DeviceName == "" {
			http.Error(w, "Missing fields", 400)
			return
		}

		var devicedEdited bool
		devicedEdited, err = EditDevice(db, deviceId, newDevice)
		if err != nil {
			http.Error(w, "internal service error", 500)
			return
		}
		if !devicedEdited {
			http.Error(w, "Device does not exist", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You updated device"))
	}

}

func DeleteDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		deviceId := req.PathValue("id")
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()

		var newDevice SmartHomeDevicePatch
		err := decoder.Decode(&newDevice)
		if err != nil {
			http.Error(w, "Improper parameters", 400)
			return
		}

		if deviceId == "" {
			http.Error(w, "Missing fields", 400)
			return
		}

		var deviceDeleted bool
		deviceDeleted, err = DeleteDevice(db, newDevice)
		if err != nil {
			http.Error(w, "internal service error", 500)
			return
		}
		if !deviceDeleted {
			http.Error(w, "Device does not exist", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("You updated device"))
	}
}

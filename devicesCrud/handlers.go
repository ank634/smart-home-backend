package devicesCrud

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	problemdetails "smart-home-backend/problemDetails"
	"strings"

	"github.com/lib/pq"
)

func AddLightDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		decoder := json.NewDecoder(req.Body)
		var newLightDevice LightDevice
		err := decoder.Decode(&newLightDevice)

		encoder := json.NewEncoder(w)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			return
		}

		err = AddLightDeviceValidator(newLightDevice)
		if err != nil {
			httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.ILLEGAL_VALUE_ERROR, Title: "title", Status: 400, Detail: "detail"}
			// header must be called before encode because w.write makes header 200 and encode calls w.write
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(httpProblemDetails)
			return
		}

		err = AddLightDevice(db, newLightDevice)
		if err != nil {
			var notNullErr ErrorNotNullViolation
			if errors.As(err, &notNullErr) {
				w.WriteHeader(http.StatusBadRequest)
				httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.NULL_NOT_ALLOWED_ERROR, Title: "No not null allowed", Status: 400, Detail: "No not null allowed"}
				w.Header().Set("Content-Type", "application/problem+json")
				encoder.Encode(httpProblemDetails)
				return
			}
			var illegalDataError ErrorIllegalData
			if errors.As(err, &illegalDataError) {
				w.WriteHeader(http.StatusBadRequest)
				httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.ILLEGAL_VALUE_ERROR, Title: "No empty strings", Status: 400, Detail: "No empty strings"}
				w.Header().Set("Content-Type", "application/problem+json")
				encoder.Encode(httpProblemDetails)
				return
			}
			var notUniqueError ErrorDuplicateData
			if errors.As(err, &notUniqueError) {
				w.WriteHeader(http.StatusBadRequest)
				httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.NULL_NOT_ALLOWED_ERROR, Title: "Values must me unique", Status: 400, Detail: "Values must me unique"}
				w.Header().Set("Content-Type", "application/problem+json")
				encoder.Encode(httpProblemDetails)
				return

			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

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
		/*
			if newDevice.DeviceID == "" ||
				newDevice.DeviceName == "" ||
				newDevice.DeviceType == "" ||
				newDevice.ServiceType == "" ||
				newDevice.DeviceUrl == "" ||
				newDevice.GetTopic == "" ||
				newDevice.SetTopic == "" {
				http.Error(w, "error: missing parameters", 400)
				return
			}*/

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

		var newDevice SmartHomeDevicePatch

		if deviceId == "" {
			http.Error(w, "Missing fields", 400)
			return
		}

		deviceDeleted, err := DeleteDevice(db, newDevice)
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

// GetDeviceHandler returns an array of Device objects as seen in models to the client
func GetDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		serviceType := req.URL.Query().Get("servicetype")
		var devices []SmartHomeDevice
		var err error
		// return array of all devices
		// TODO:maybe we should type casting to enums to secure proper service type
		if serviceType == "" {
			devices, err = GetDevicesByServiceType(db, serviceType)
		} else {
			devices, err = GetAllDevices(db)
		}

		if err != nil {
			http.Error(w, "error: internal server error", http.StatusInternalServerError)
			w.Write([]byte("error: could not fetch devices"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		// if encode is sucessful it writes to the writer
		err = encoder.Encode(devices)
		if err != nil {
			http.Error(w, "error: internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// NOTE: I need to learn more idiomatic go it may be more appropriate to return a struct
func AddLightDeviceValidator(light LightDevice) error {
	if light.DeviceID == nil ||
		light.DeviceName == nil ||
		light.DeviceType == nil ||
		light.EndPoint == nil ||
		light.GetTopic == nil ||
		light.IsDimmable == nil ||
		light.IsRgb == nil ||
		light.Manufactor == nil ||
		light.ServiceType == nil ||
		light.SetTopic == nil {
		return ErrorIllegalData{fmt.Sprint("All fields except room number may not be nil")}
	}

	if strings.TrimSpace(*light.DeviceID) == "" ||
		strings.TrimSpace(*light.DeviceName) == "" ||
		strings.TrimSpace(*light.DeviceType) == "" ||
		strings.TrimSpace(*light.EndPoint) == "" ||
		strings.TrimSpace(*light.GetTopic) == "" ||
		strings.TrimSpace(*light.Manufactor) == "" ||
		strings.TrimSpace(*light.ServiceType) == "" ||
		strings.TrimSpace(*light.SetTopic) == "" {
		return ErrorIllegalData{fmt.Sprint("All fields except room number may not be nil")}
	}
	return nil
}

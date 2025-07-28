package devicesCrud

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	problemdetails "smart-home-backend/problemDetails"
	"strings"
)

// TODO: give a better error message for when roomid does not exist
func AddDevice(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		requestBodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		// copy the contents into to seperate streams so can decode twice later on
		req.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
		requestBodyCopy := io.NopCloser(bytes.NewBuffer(requestBodyBytes))

		var device SmartHomeDevice
		err = json.NewDecoder(req.Body).Decode(&device)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = AddDeviceValidator(device)
		if err != nil {
			var errNull ErrorNotNullViolation
			var errIllegalData ErrorIllegalData
			if errors.As(err, &errNull) {
				problemdetails.ProblemDetail(w, problemdetails.NULL_NOT_ALLOWED_ERROR, "Null is not allowed", http.StatusBadRequest, "Null is not allowed")
				return
			}
			if errors.As(err, &errIllegalData) {
				problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Empty strings are not allowed", http.StatusBadRequest, "Empty strings not allowed")
				return
			}
		}

		if strings.ToLower(*device.DeviceType) == "light" {
			var light LightDevice
			req.Body = requestBodyCopy
			err = json.NewDecoder(req.Body).Decode(&light)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			err = AddLightDevice(db, light)
		} else {
			problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Device type is not supported", 400, "Device type is not supported")
		}

		if err != nil {
			var notNullErr ErrorNotNullViolation
			if errors.As(err, &notNullErr) {
				problemdetails.ProblemDetail(w, problemdetails.NULL_NOT_ALLOWED_ERROR, "Null not allowed", http.StatusBadRequest, "Null not allowed")
				return
			}
			var illegalDataError ErrorIllegalData
			if errors.As(err, &illegalDataError) {
				problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Value not allowed", http.StatusBadRequest, "Value not allowed")
				return
			}
			var notUniqueError ErrorDuplicateData
			if errors.As(err, &notUniqueError) {
				problemdetails.ProblemDetail(w, problemdetails.NOT_UNIQUE_ERROR, "non unique value not allowed", http.StatusBadRequest, "non unique value not allowed")
				return

			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// Done
func EditDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		deviceId := req.PathValue("id")
		decoder := json.NewDecoder(req.Body)

		var newDevice SmartHomeDevicePatch
		err := decoder.Decode(&newDevice)
		if err != nil {
			http.Error(w, "internal service error", 500)
			return
		}

		if strings.TrimSpace(deviceId) == "" || strings.TrimSpace(newDevice.DeviceName) == "" {
			problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Empty strings are not valid values", 400, "Empty strings are not valid values")
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

// DONE
func DeleteDeviceHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		deviceId := req.PathValue("id")

		if strings.TrimSpace(deviceId) == "" {
			problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Empty strings are not valid values", 400, "Empty strings are not valid values")
			return
		}

		deviceDeleted, err := DeleteDevice(db, deviceId)
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
		return ErrorNotNullViolation{fmt.Sprint("All fields except room number may not be nil")}
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

// NOTE: I need to learn more idiomatic go it may be more appropriate to return a struct
func AddDeviceValidator(device SmartHomeDevice) error {
	if device.DeviceID == nil ||
		device.DeviceName == nil ||
		device.DeviceType == nil ||
		device.EndPoint == nil ||
		device.GetTopic == nil ||
		device.Manufactor == nil ||
		device.ServiceType == nil ||
		device.SetTopic == nil {
		return ErrorNotNullViolation{fmt.Sprint("All fields except room number may not be nil")}
	}

	if strings.TrimSpace(*device.DeviceID) == "" ||
		strings.TrimSpace(*device.DeviceName) == "" ||
		strings.TrimSpace(*device.DeviceType) == "" ||
		strings.TrimSpace(*device.EndPoint) == "" ||
		strings.TrimSpace(*device.GetTopic) == "" ||
		strings.TrimSpace(*device.Manufactor) == "" ||
		strings.TrimSpace(*device.ServiceType) == "" ||
		strings.TrimSpace(*device.SetTopic) == "" {
		return ErrorIllegalData{fmt.Sprint("All fields except room number may not be nil")}
	}
	return nil
}

func AddRoomHandler(db *sql.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var room Room
		err := json.NewDecoder(req.Body).Decode(&room)
		if err != nil {
			http.Error(w, "internal server error", 500)
		}
		if room.RoomName == nil {
			problemdetails.ProblemDetail(w, problemdetails.NULL_NOT_ALLOWED_ERROR, "Null not allowed", http.StatusBadRequest, "Null not allowed")
			return
		}
		if strings.TrimSpace(*room.RoomName) == "" {
			problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Value not allowed", http.StatusBadRequest, "Value not allowed")
			return
		}

		err = AddRoom(db, *room.RoomName)
		if err != nil {
			var notNullErr ErrorNotNullViolation
			if errors.As(err, &notNullErr) {
				problemdetails.ProblemDetail(w, problemdetails.NULL_NOT_ALLOWED_ERROR, "Null not allowed", http.StatusBadRequest, "Null not allowed")
				return
			}
			var illegalDataError ErrorIllegalData
			if errors.As(err, &illegalDataError) {
				problemdetails.ProblemDetail(w, problemdetails.ILLEGAL_VALUE_ERROR, "Value not allowed", http.StatusBadRequest, "Value not allowed")
				return
			}
			var notUniqueError ErrorDuplicateData
			if errors.As(err, &notUniqueError) {
				problemdetails.ProblemDetail(w, problemdetails.NOT_UNIQUE_ERROR, "non unique value not allowed", http.StatusBadRequest, "non unique value not allowed")
				return

			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

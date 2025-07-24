package devicesCrud

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	problemdetails "smart-home-backend/problemDetails"
	"strings"
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

		encoder := json.NewEncoder(w)
		if strings.TrimSpace(deviceId) == "" || strings.TrimSpace(newDevice.DeviceName) == "" || strings.TrimSpace(newDevice.DeviceID) == "" {
			w.WriteHeader(http.StatusBadRequest)
			httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.ILLEGAL_VALUE_ERROR, Title: "No empty strings", Status: 400, Detail: "No empty strings"}
			w.Header().Set("Content-Type", "application/problem+json")
			encoder.Encode(httpProblemDetails)
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

		encoder := json.NewEncoder(w)
		if strings.TrimSpace(deviceId) == "" {
			w.WriteHeader(http.StatusBadRequest)
			httpProblemDetails := problemdetails.ProblemDetail{ErrorType: problemdetails.ILLEGAL_VALUE_ERROR, Title: "No empty strings", Status: 400, Detail: "No empty strings"}
			w.Header().Set("Content-Type", "application/problem+json")
			encoder.Encode(httpProblemDetails)
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

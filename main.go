package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type ApiResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    []Appointment `json:"data"`
}

type ApiError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type Appointment struct {
	BarberID   int       `json:"barber_id"`
	CustomerID int       `json:"customer_id"`
	ServiceID  int       `json:"service_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

var appointments []Appointment

func IsBarberAvailable(barberID int, desiredTime time.Time, duration time.Duration) bool {
	desiredEndTime := desiredTime.Add(duration)
	for _, appointment := range appointments {
		if appointment.BarberID == barberID {
			if desiredTime.Before(appointment.EndTime) && desiredEndTime.After(appointment.StartTime) {
				return false
			}
		}
	}
	return true
}

func ScheduleAppointment(barberID, customerID, serviceID int, desiredTime time.Time, duration time.Duration) (*Appointment, *ApiError) {
	if IsBarberAvailable(barberID, desiredTime, duration) {
		newAppointment := Appointment{
			BarberID:   barberID,
			CustomerID: customerID,
			ServiceID:  serviceID,
			StartTime:  desiredTime,
			EndTime:    desiredTime.Add(duration),
		}
		appointments = append(appointments, newAppointment)
		return &newAppointment, nil
	} else {
		return nil, &ApiError{Error: true, Message: "O barbeiro não está disponível no horário desejado."}
	}
}

func CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var newAppointment Appointment
	err := json.NewDecoder(r.Body).Decode(&newAppointment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appointment, apiError := ScheduleAppointment(newAppointment.BarberID, newAppointment.CustomerID, newAppointment.ServiceID, newAppointment.StartTime, 30*time.Minute)
	if apiError != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(apiError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(appointment)
}

func GetAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	response := ApiResponse{
		Success: true,
		Message: "Lista de agendamentos recuperada com sucesso.",
		Data:    appointments,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/agendamentos", CreateAppointmentHandler).Methods("POST")
	router.HandleFunc("/agendamentos", GetAppointmentsHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

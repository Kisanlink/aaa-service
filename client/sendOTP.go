package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// SMSRequest represents the JSON payload for sending an SMS.
type SMSRequest struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
}

// SendOTP sends an OTP code to the given mobile phone number by calling the notification service HTTP endpoint.
func SendOTP(countryCode, phoneNumber, otp string) {
	// Create an HTTP client with a timeout.
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare the JSON payload.
	requestPayload := SMSRequest{
		PhoneNumber: countryCode + phoneNumber,
		Message:     fmt.Sprintf("Your OTP is: %s", otp),
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Failed to marshal SMS request: %v", err)
		return
	}

	// Build the URL from an environment variable.
	url := fmt.Sprintf("%s/api/v1/send/sms", os.Getenv("NOTIFICATION_SERVICE_ENDPOINT"))

	// Send the HTTP POST request.
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send OTP SMS request: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Println(resp)
	// Read and log the response from the notification service.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read OTP SMS response: %v", err)
		return
	}

	log.Printf("OTP sent successfully. Response: %s", string(body))
}

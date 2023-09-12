package handlers

import (
	"github.com/fairDataSociety/FaVe/models"
)

// createErrorResponseObject is a common function to create an error response
func createErrorResponseObject(messages ...string) *models.ErrorResponse {
	// Initialize return value
	er := &models.ErrorResponse{}

	// appends all error messages to the error
	for _, message := range messages {
		er.Error = append(er.Error, &models.ErrorResponseErrorItems0{
			Message: message,
		})
	}

	return er
}

// createOKResponseObject is a common function to create an ok response
func createOKResponseObject(messages string) *models.OKResponse {
	return &models.OKResponse{
		Message: messages,
	}
}

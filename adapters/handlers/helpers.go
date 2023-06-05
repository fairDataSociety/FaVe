package handlers

import (
	"fmt"
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

func errPayloadFromSingleErr(err error) *models.ErrorResponse {
	return &models.ErrorResponse{Error: []*models.ErrorResponseErrorItems0{{
		Message: fmt.Sprintf("%s", err),
	}}}
}

// createOKResponseObject is a common function to create an ok response
func createOKResponseObject(messages string) *models.OKResponse {
	return &models.OKResponse{
		Message: messages,
	}
}

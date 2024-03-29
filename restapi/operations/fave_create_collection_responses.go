// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/fairDataSociety/FaVe/models"
)

// FaveCreateCollectionOKCode is the HTTP code returned for type FaveCreateCollectionOK
const FaveCreateCollectionOKCode int = 200

/*
FaveCreateCollectionOK collection added

swagger:response faveCreateCollectionOK
*/
type FaveCreateCollectionOK struct {

	/*
	  In: Body
	*/
	Payload *models.OKResponse `json:"body,omitempty"`
}

// NewFaveCreateCollectionOK creates FaveCreateCollectionOK with default headers values
func NewFaveCreateCollectionOK() *FaveCreateCollectionOK {

	return &FaveCreateCollectionOK{}
}

// WithPayload adds the payload to the fave create collection o k response
func (o *FaveCreateCollectionOK) WithPayload(payload *models.OKResponse) *FaveCreateCollectionOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the fave create collection o k response
func (o *FaveCreateCollectionOK) SetPayload(payload *models.OKResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *FaveCreateCollectionOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// FaveCreateCollectionBadRequestCode is the HTTP code returned for type FaveCreateCollectionBadRequest
const FaveCreateCollectionBadRequestCode int = 400

/*
FaveCreateCollectionBadRequest Malformed request.

swagger:response faveCreateCollectionBadRequest
*/
type FaveCreateCollectionBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewFaveCreateCollectionBadRequest creates FaveCreateCollectionBadRequest with default headers values
func NewFaveCreateCollectionBadRequest() *FaveCreateCollectionBadRequest {

	return &FaveCreateCollectionBadRequest{}
}

// WithPayload adds the payload to the fave create collection bad request response
func (o *FaveCreateCollectionBadRequest) WithPayload(payload *models.ErrorResponse) *FaveCreateCollectionBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the fave create collection bad request response
func (o *FaveCreateCollectionBadRequest) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *FaveCreateCollectionBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// FaveCreateCollectionUnauthorizedCode is the HTTP code returned for type FaveCreateCollectionUnauthorized
const FaveCreateCollectionUnauthorizedCode int = 401

/*
FaveCreateCollectionUnauthorized Unauthorized or invalid credentials.

swagger:response faveCreateCollectionUnauthorized
*/
type FaveCreateCollectionUnauthorized struct {
}

// NewFaveCreateCollectionUnauthorized creates FaveCreateCollectionUnauthorized with default headers values
func NewFaveCreateCollectionUnauthorized() *FaveCreateCollectionUnauthorized {

	return &FaveCreateCollectionUnauthorized{}
}

// WriteResponse to the client
func (o *FaveCreateCollectionUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// FaveCreateCollectionForbiddenCode is the HTTP code returned for type FaveCreateCollectionForbidden
const FaveCreateCollectionForbiddenCode int = 403

/*
FaveCreateCollectionForbidden Forbidden

swagger:response faveCreateCollectionForbidden
*/
type FaveCreateCollectionForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewFaveCreateCollectionForbidden creates FaveCreateCollectionForbidden with default headers values
func NewFaveCreateCollectionForbidden() *FaveCreateCollectionForbidden {

	return &FaveCreateCollectionForbidden{}
}

// WithPayload adds the payload to the fave create collection forbidden response
func (o *FaveCreateCollectionForbidden) WithPayload(payload *models.ErrorResponse) *FaveCreateCollectionForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the fave create collection forbidden response
func (o *FaveCreateCollectionForbidden) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *FaveCreateCollectionForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// FaveCreateCollectionUnprocessableEntityCode is the HTTP code returned for type FaveCreateCollectionUnprocessableEntity
const FaveCreateCollectionUnprocessableEntityCode int = 422

/*
FaveCreateCollectionUnprocessableEntity Request body is well-formed (i.e., syntactically correct), but semantically erroneous.

swagger:response faveCreateCollectionUnprocessableEntity
*/
type FaveCreateCollectionUnprocessableEntity struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewFaveCreateCollectionUnprocessableEntity creates FaveCreateCollectionUnprocessableEntity with default headers values
func NewFaveCreateCollectionUnprocessableEntity() *FaveCreateCollectionUnprocessableEntity {

	return &FaveCreateCollectionUnprocessableEntity{}
}

// WithPayload adds the payload to the fave create collection unprocessable entity response
func (o *FaveCreateCollectionUnprocessableEntity) WithPayload(payload *models.ErrorResponse) *FaveCreateCollectionUnprocessableEntity {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the fave create collection unprocessable entity response
func (o *FaveCreateCollectionUnprocessableEntity) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *FaveCreateCollectionUnprocessableEntity) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(422)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// FaveCreateCollectionInternalServerErrorCode is the HTTP code returned for type FaveCreateCollectionInternalServerError
const FaveCreateCollectionInternalServerErrorCode int = 500

/*
FaveCreateCollectionInternalServerError An error has occurred while trying to fulfill the request. Most likely the ErrorResponse will contain more information about the error.

swagger:response faveCreateCollectionInternalServerError
*/
type FaveCreateCollectionInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewFaveCreateCollectionInternalServerError creates FaveCreateCollectionInternalServerError with default headers values
func NewFaveCreateCollectionInternalServerError() *FaveCreateCollectionInternalServerError {

	return &FaveCreateCollectionInternalServerError{}
}

// WithPayload adds the payload to the fave create collection internal server error response
func (o *FaveCreateCollectionInternalServerError) WithPayload(payload *models.ErrorResponse) *FaveCreateCollectionInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the fave create collection internal server error response
func (o *FaveCreateCollectionInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *FaveCreateCollectionInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

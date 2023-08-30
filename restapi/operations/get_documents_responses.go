// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/fairDataSociety/FaVe/models"
)

// GetDocumentsOKCode is the HTTP code returned for type GetDocumentsOK
const GetDocumentsOKCode int = 200

/*
GetDocumentsOK Successful response

swagger:response getDocumentsOK
*/
type GetDocumentsOK struct {

	/*
	  In: Body
	*/
	Payload *models.Document `json:"body,omitempty"`
}

// NewGetDocumentsOK creates GetDocumentsOK with default headers values
func NewGetDocumentsOK() *GetDocumentsOK {

	return &GetDocumentsOK{}
}

// WithPayload adds the payload to the get documents o k response
func (o *GetDocumentsOK) WithPayload(payload *models.Document) *GetDocumentsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get documents o k response
func (o *GetDocumentsOK) SetPayload(payload *models.Document) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDocumentsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetDocumentsBadRequestCode is the HTTP code returned for type GetDocumentsBadRequest
const GetDocumentsBadRequestCode int = 400

/*
GetDocumentsBadRequest Malformed request.

swagger:response getDocumentsBadRequest
*/
type GetDocumentsBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewGetDocumentsBadRequest creates GetDocumentsBadRequest with default headers values
func NewGetDocumentsBadRequest() *GetDocumentsBadRequest {

	return &GetDocumentsBadRequest{}
}

// WithPayload adds the payload to the get documents bad request response
func (o *GetDocumentsBadRequest) WithPayload(payload *models.ErrorResponse) *GetDocumentsBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get documents bad request response
func (o *GetDocumentsBadRequest) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDocumentsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetDocumentsNotFoundCode is the HTTP code returned for type GetDocumentsNotFound
const GetDocumentsNotFoundCode int = 404

/*
GetDocumentsNotFound Document not found

swagger:response getDocumentsNotFound
*/
type GetDocumentsNotFound struct {
}

// NewGetDocumentsNotFound creates GetDocumentsNotFound with default headers values
func NewGetDocumentsNotFound() *GetDocumentsNotFound {

	return &GetDocumentsNotFound{}
}

// WriteResponse to the client
func (o *GetDocumentsNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
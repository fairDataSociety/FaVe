// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetDocumentsHandlerFunc turns a function with the right signature into a get documents handler
type GetDocumentsHandlerFunc func(GetDocumentsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetDocumentsHandlerFunc) Handle(params GetDocumentsParams) middleware.Responder {
	return fn(params)
}

// GetDocumentsHandler interface for that can handle valid get documents params
type GetDocumentsHandler interface {
	Handle(GetDocumentsParams) middleware.Responder
}

// NewGetDocuments creates a new http.Handler for the get documents operation
func NewGetDocuments(ctx *middleware.Context, handler GetDocumentsHandler) *GetDocuments {
	return &GetDocuments{Context: ctx, Handler: handler}
}

/*
	GetDocuments swagger:route GET /documents getDocuments

Retrieve a document based on query parameters
*/
type GetDocuments struct {
	Context *middleware.Context
	Handler GetDocumentsHandler
}

func (o *GetDocuments) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetDocumentsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
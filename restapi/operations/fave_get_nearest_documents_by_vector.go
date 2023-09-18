// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// FaveGetNearestDocumentsByVectorHandlerFunc turns a function with the right signature into a fave get nearest documents by vector handler
type FaveGetNearestDocumentsByVectorHandlerFunc func(FaveGetNearestDocumentsByVectorParams) middleware.Responder

// Handle executing the request and returning a response
func (fn FaveGetNearestDocumentsByVectorHandlerFunc) Handle(params FaveGetNearestDocumentsByVectorParams) middleware.Responder {
	return fn(params)
}

// FaveGetNearestDocumentsByVectorHandler interface for that can handle valid fave get nearest documents by vector params
type FaveGetNearestDocumentsByVectorHandler interface {
	Handle(FaveGetNearestDocumentsByVectorParams) middleware.Responder
}

// NewFaveGetNearestDocumentsByVector creates a new http.Handler for the fave get nearest documents by vector operation
func NewFaveGetNearestDocumentsByVector(ctx *middleware.Context, handler FaveGetNearestDocumentsByVectorHandler) *FaveGetNearestDocumentsByVector {
	return &FaveGetNearestDocumentsByVector{Context: ctx, Handler: handler}
}

/*
	FaveGetNearestDocumentsByVector swagger:route POST /nearest-documents-by-vector faveGetNearestDocumentsByVector

Get nearest documents for a collection.
*/
type FaveGetNearestDocumentsByVector struct {
	Context *middleware.Context
	Handler FaveGetNearestDocumentsByVectorHandler
}

func (o *FaveGetNearestDocumentsByVector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewFaveGetNearestDocumentsByVectorParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

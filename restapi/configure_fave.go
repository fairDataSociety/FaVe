// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fairDataSociety/FaVe"
	"github.com/fairDataSociety/FaVe/adapters/handlers"
	"github.com/fairDataSociety/FaVe/restapi/operations"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
)

//go:generate swagger generate server --target ../../FaVe --name Fave --spec ../openapi-spec/schema.json --principal interface{}

func configureFlags(api *operations.FaveAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.FaveAPI) http.Handler {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	handlerConfig := handlers.HandlerConfig{}
	handler, err := handlers.NewHandler(ctx, &handlerConfig)
	if err != nil {
		fmt.Println("Error creating handler: ", err)
		os.Exit(1)
	}
	api.FaveRootHandler = operations.FaveRootHandlerFunc(handler.FaveRootHandler)
	fmt.Println("Version: ", FaVe.Version)
	api.FaveCreateCollectionHandler = operations.FaveCreateCollectionHandlerFunc(handler.FaveCreateCollectionHandler)
	api.FaveGetCollectionsHandler = operations.FaveGetCollectionsHandlerFunc(handler.FaveGetCollectionsHandler)
	api.FaveDeleteCollectionHandler = operations.FaveDeleteCollectionHandlerFunc(handler.FaveDeleteCollectionHandler)
	api.FaveAddDocumentsHandler = operations.FaveAddDocumentsHandlerFunc(handler.FaveAddDocumentsHandler)
	api.FaveGetDocumentsHandler = operations.FaveGetDocumentsHandlerFunc(handler.GetDocumentsHandlerFunc)
	api.FaveGetNearestDocumentsHandler = operations.FaveGetNearestDocumentsHandlerFunc(handler.FaveGetNearestDocumentsHandler)
	api.FaveGetNearestDocumentsByVectorHandler = operations.FaveGetNearestDocumentsByVectorHandlerFunc(handler.FaveGetNearestDocumentsByVectorHandler)
	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}

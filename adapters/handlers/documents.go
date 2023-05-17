package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/fairDataSociety/FaVe/models"
	"github.com/fairDataSociety/FaVe/pkg/document"
	"github.com/fairDataSociety/FaVe/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

func (s *Handler) FaveAddDocumentsHandler(request operations.FaveAddDocumentsParams) middleware.Responder {
	documentsRaw := request.Body
	if documentsRaw.Name == "" {
		return operations.NewFaveAddDocumentsBadRequest().WithPayload(createErrorResponseObject("Collection name cannot be blank"))
	}
	if documentsRaw.Documents == nil {
		return operations.NewFaveAddDocumentsBadRequest().WithPayload(createErrorResponseObject("Request should have at least one document"))
	}
	documents := make([]*document.Document, len(documentsRaw.Documents))

	for i, v := range documentsRaw.Documents {
		props, ok := v.Properties.(map[string]interface{})
		if !ok {
			continue
		}
		d := &document.Document{
			ID:         string(v.ID),
			Properties: props,
		}
		documents[i] = d
	}

	err := s.doc.AddDocuments(documentsRaw.Name, documents...)
	if err != nil {
		return operations.NewFaveAddDocumentsBadRequest().WithPayload(createErrorResponseObject("Failed to create add documents"))
	}
	return operations.NewFaveAddDocumentsOK()
}

func (s *Handler) FaveGetNearestDocumentsHandler(request operations.FaveGetNearestDocumentsParams) middleware.Responder {
	req := request.Body
	if req.Name == "" {
		return operations.NewFaveGetNearestDocumentsBadRequest().WithPayload(createErrorResponseObject("Collection name cannot be blank"))
	}
	if req.Text == "" {
		return operations.NewFaveGetNearestDocumentsBadRequest().WithPayload(createErrorResponseObject("Search text should not be blank"))
	}
	documentsRaw, err := s.doc.GetNearDocuments(req.Name, req.Text, req.Distance)
	if err != nil {
		return operations.NewFaveGetNearestDocumentsBadRequest().WithPayload(createErrorResponseObject("Failed to get nearest documents :" + err.Error()))
	}

	documents := make([]*models.Document, len(documentsRaw))
	for i, v := range documentsRaw {
		props := map[string]interface{}{}
		err := json.Unmarshal(v, &props)
		if err != nil {
			return operations.NewFaveGetNearestDocumentsBadRequest().WithPayload(createErrorResponseObject("Failed to get nearest documents :" + err.Error()))
		}
		d := &models.Document{
			ID: strfmt.UUID(fmt.Sprintf("%s", props["id"])),
		}
		delete(props, "id")
		delete(props, "vector")
		d.Properties = props
		documents[i] = d
	}

	return operations.NewFaveGetNearestDocumentsOK().WithPayload(&models.NearestDocumentsResponse{
		Documents: documents,
		Name:      req.Name,
	})
}

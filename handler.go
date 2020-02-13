package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"log"
	"net/http"
)

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	jsonContentType       = "application/json"
)

type PatchHandler struct{}

func NewPatchHandler() *PatchHandler {
	return &PatchHandler{}
}

func (p *PatchHandler) generatePatchOperations() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Handling webhook request ...")

		var writeErr error
		if bytes, err := doPatchOperationGeneratorFunc(w, r); err != nil {
			log.Printf("Error handling webhook request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr = w.Write([]byte(err.Error()))
		} else {
			log.Print("Webhook request handled successfully")
			_, writeErr = w.Write(bytes)
		}

		if writeErr != nil {
			log.Printf("Could not write response: %v", writeErr)
		}
	})
}

func doPatchOperationGeneratorFunc(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not read request body: %v", err)
	}

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("unsupported content type %s, only %s is supported", contentType, jsonContentType)
	}

	admissionReviewReq := &v1beta1.AdmissionReview{}

	if _, _, err := universalDeserializer.Decode(body, nil, admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("malformed admission review: request is nil")
	}

	admissionReviewResp := &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: true,
			UID:     admissionReviewReq.Request.UID,
		},
	}

	var patchOperations []patchOperation

	if !isKubeNamespace(admissionReviewReq.Request.Namespace) {
		patchOperations, err = generatePodPatches(admissionReviewReq.Request)
	}

	if err != nil {
		admissionReviewResp.Response.Allowed = false
		admissionReviewResp.Response.Result = &metav1.Status{
			Message: err.Error(),
		}
	} else {
		patchOperationsBytes, err := json.Marshal(patchOperations)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil, fmt.Errorf("could not marshal JSON patch: %v", err)
		}

		admissionReviewResp.Response.Allowed = true
		admissionReviewResp.Response.Patch = patchOperationsBytes
	}

	bytes, err := json.Marshal(&admissionReviewResp)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %v", err)
	}
	return bytes, nil
}

func isKubeNamespace(ns string) bool {
	return ns == metav1.NamespacePublic || ns == metav1.NamespaceSystem
}

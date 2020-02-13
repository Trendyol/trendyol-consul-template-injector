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

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type patchOperationGeneratorFunc func(*v1beta1.AdmissionRequest) ([]patchOperation, error)

func patchHttpHandler(m patchOperationGeneratorFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		patchOperationGeneratorFuncWrapper(w, r, m)
	})
}

func patchOperationGeneratorFuncWrapper(w http.ResponseWriter, r *http.Request, m patchOperationGeneratorFunc) {
	log.Print("Handling webhook request ...")

	var writeErr error
	if bytes, err := executePatchOperationGeneratorFunc(w, r, m); err != nil {
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
}

func executePatchOperationGeneratorFunc(w http.ResponseWriter, r *http.Request, mf patchOperationGeneratorFunc) ([]byte, error) {
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
		patchOperations, err = mf(admissionReviewReq.Request)
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

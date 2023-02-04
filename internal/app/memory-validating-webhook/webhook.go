package main

import (
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

type WebhookServer struct {
	server *http.Server
}

func (webhookServer *WebhookServer) dispatch(response http.ResponseWriter, request *http.Request) {
	var body []byte
	if request.Body != nil {
		if data, err := ioutil.ReadAll(request.Body); err == nil {
			body = data
		}
	}
	if body == nil {
		http.Error(response, "request body is empty", http.StatusBadRequest)
		return
	}
	glog.Infoln(string(body))
	admissionReview := v1.AdmissionReview{}
	var admissionResponse *v1.AdmissionResponse
	if _, _, err := deserializer.Decode(body, nil, &admissionReview); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if request.URL.Path == "/mutate" {
			admissionResponse = webhookServer.mutate(&admissionReview)
		} else if request.URL.Path == "/validate" {
			admissionResponse = webhookServer.validate(&admissionReview)
		}
	}

	admissionReview.Response = admissionResponse
	responseBytes, _ := json.Marshal(admissionReview)
	glog.Infoln(string(responseBytes))

	response.WriteHeader(200)
	response.Header().Add("Content-Type", "application/json")
	response.Write(responseBytes)
}

func (webhookServer WebhookServer) mutate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	response := &v1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}
	return response
}

func (webhookServer *WebhookServer) validate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	response := &v1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}
	return response
}

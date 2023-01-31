package main

import (
	"fmt"
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
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		body = data
	}

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
		fmt.Println(request.URL.Path)
		if request.URL.Path == "/mutate" {
			admissionResponse = webhookServer.mutate(&admissionReview)
		} else if request.URL.Path == "/validate" {
			admissionResponse = webhookServer.validate(&admissionReview)
		}
	}
	responseBytes, err := json.Marshal(&admissionResponse)
	response.WriteHeader(200)
	response.Header().Add("Content-Type", "application/json")
	response.Write(responseBytes)
}

func (webhookServer WebhookServer) mutate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	return nil
}

func (webhookServer *WebhookServer) validate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	return nil
}
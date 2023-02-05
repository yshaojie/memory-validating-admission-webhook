package memory_validating_webhook

import (
	"github.com/golang/glog"
	"github.com/wI2L/jsondiff"
	"io/ioutil"
	"k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	Server *http.Server
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (webhookServer *WebhookServer) Dispatch(response http.ResponseWriter, request *http.Request) {
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
	response.WriteHeader(http.StatusOK)
	response.Header().Add("Content-Type", "application/json")
	response.Write(responseBytes)
}

func (webhookServer WebhookServer) mutate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	var patchBytes []byte
	request := admissionReview.Request
	switch request.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(request.Object.Raw, &deployment); err != nil {
			return &v1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		deploymentCopy := deployment.DeepCopy()
		mutateDeployment(deploymentCopy)
		patch, _ := jsondiff.Compare(deployment, deploymentCopy)
		patchb, _ := json.Marshal(patch)
		patchBytes = patchb
	case "Pod":

	}
	return &v1.AdmissionResponse{
		UID:     request.UID,
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func mutateDeployment(deploymentCopy *appsv1.Deployment) error {
	containers := deploymentCopy.Spec.Template.Spec.Containers

	for i, container := range containers {
		envVar := corev1.EnvVar{
			Name: "K8S_POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		}
		containers[i].Env = append(container.Env, envVar)
	}
	aa, _ := json.Marshal(containers)
	glog.Infoln(string(aa))
	bytes, _ := json.Marshal(deploymentCopy)
	glog.Infoln(string(bytes))
	return nil
}

func (webhookServer *WebhookServer) validate(admissionReview *v1.AdmissionReview) *v1.AdmissionResponse {
	response := &v1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}
	return response
}

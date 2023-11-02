// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package mutate

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/rs/zerolog/log"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mutate mutates
func Mutate(body []byte) ([]byte, error) {
	log.Debug().Msg(fmt.Sprintf("recv: %s\n", string(body)))

	admReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	log.Debug().Msg(fmt.Sprintf("admission review: %v\n", admReview))

	pod, err := getPod(admReview.Request)
	if err != nil {
		return nil, fmt.Errorf("error extracting pod: %w", err)
	}

	patchBytes, err := createPatchForPodAnnotation(admReview.Request.Object.Raw, pod)
	if err != nil {
		return nil, fmt.Errorf("error creating pod annotation patch: %w", err)
	}

	admReview.Response = &v1.AdmissionResponse{
		Patch: patchBytes,
		Result: &metav1.Status{
			Status: "Success",
		},
		AuditAnnotations: map[string]string{
			"assignedBy": "devanggaur",
		},
		Allowed: true,
		UID:     admReview.Request.UID,
		PatchType: func() *v1.PatchType {
			if len(patchBytes) == 0 {
				return nil
			}
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	responseBody, err := json.Marshal(admReview)
	if err != nil {
		return nil, fmt.Errorf("error marshaling admission response: %v", err)
	}

	return responseBody, nil
}

func addAnnotation(pod *corev1.Pod) (list []byte, err error) {
	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = make(map[string]string)
	}

	pod.ObjectMeta.Annotations["owner"] = "devang"
	return json.Marshal(pod)
}

func getPod(admReq *v1.AdmissionRequest) (*corev1.Pod, error) {
	var pod *corev1.Pod

	log.Debug().Msg(fmt.Sprintf("admission review request: %v\n", admReq))

	if admReq == nil {
		return nil, fmt.Errorf("admission request is nil")
	}

	if err := json.Unmarshal(admReq.Object.Raw, &pod); err != nil {
		return nil, fmt.Errorf("unable to unmarshal pod json object %v", err)
	}

	return pod, nil
}

func createPatchForPodAnnotation(rawObject []byte, pod *corev1.Pod) ([]byte, error) {
	annotatedPodBytes, err := addAnnotation(pod)
	if err != nil {
		return nil, fmt.Errorf("error adding annotation: %w", err)
	}

	patches, err := jsonpatch.CreatePatch(rawObject, annotatedPodBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch: %w", err)
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patches")
	}

	return patchBytes, nil
}

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
func Mutate(body []byte, verbose bool) ([]byte, error) {
	if verbose {
		log.Info().Msg(fmt.Sprintf("recv: %s\n", string(body)))
	}

	// unmarshal request into AdmissionReview struct
	admReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	if verbose {
		log.Info().Msg(fmt.Sprintf("admission review: %v\n", admReview))
	}

	var pod *corev1.Pod

	ar := admReview.Request

	if verbose {
		log.Info().Msg(fmt.Sprintf("admission review request: %v\n", ar))
	}

	if ar == nil {
		return nil, fmt.Errorf("admission request is nil")
	}

	// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
	if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
		return nil, fmt.Errorf("unable to unmarshal pod json object %v", err)
	}

	if verbose {
		log.Info().Msg(fmt.Sprintf("pod: %v\n", pod))
	}

	var err error

	annotatedPodBytes, err := addAnnotation(pod)
	if err != nil {
		return nil, fmt.Errorf("error adding annotation: %v", err)
	}

	if verbose {
		log.Info().Msg(fmt.Sprintf("annotated pod: %s\n", string(annotatedPodBytes)))
	}

	patches, err := jsonpatch.CreatePatch(ar.Object.Raw, annotatedPodBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch: %v", err)
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patches")
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
		UID:     ar.UID,
		PatchType: func() *v1.PatchType {
			if len(patches) == 0 {
				return nil
			}
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err := json.Marshal(admReview)
	if err != nil {
		return nil, fmt.Errorf("error marshaling admission response: %v", err)
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody)) // untested section
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

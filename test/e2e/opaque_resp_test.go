package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOpaqueResp(t *testing.T) {
	coreClient := coreClientset(t)

	clearConfigMapsInAllNses(t, coreClient, nses)

	for _, ns := range nses {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: ns,
			},
			Data: map[string]string{},
		}

		timeStep("config map listing (per ns)", func() {
			_, err := coreClient.CoreV1().ConfigMaps(ns).Create(cm)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})

		timeStep("config map listing (per ns)", func() {
			cmList, err := coreClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf(err.Error())
			}

			if len(cmList.Items) != 1 {
				t.Fatalf("Expected number of config maps to be 1 but was %d", len(cmList.Items))
			}
		})
	}
}

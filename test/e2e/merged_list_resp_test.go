package e2e

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergedListResp(t *testing.T) {
	coreClient := coreClientset(t)

	clearConfigMapsInAllNses(t, coreClient, nses)

	timeStep("config map creation (all ns)", func() {
		for _, ns := range nses {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: ns,
				},
				Data: map[string]string{},
			}

			_, err := coreClient.CoreV1().ConfigMaps(ns).Create(cm)
			if err != nil {
				t.Fatalf(err.Error())
			}
		}
	})

	timeStep("config map listing (empty ns)", func() {
		cmList, err := coreClient.CoreV1().ConfigMaps("").List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf(err.Error())
		}

		// not checking exact number because tests can run under Cluster level perms
		// which will pull in all config maps across all namespaces
		if len(cmList.Items) < len(nses) {
			t.Fatalf("Expected number of config maps to be <%d but was %d", len(nses), len(cmList.Items))
		}

		fmt.Printf("got %d config maps\n", len(cmList.Items))
	})
}

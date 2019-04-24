package e2e

import (
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var nses = []string{"rbac-proxy-test-ns1", "rbac-proxy-test-ns2", "rbac-proxy-test-ns3"}

func coreClientset(t *testing.T) kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// TODO high QPS
	config.QPS = 100000
	config.Burst = 100000

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return clientset
}

func clearConfigMapsInAllNses(t *testing.T, coreClient kubernetes.Interface, nses []string) {
	for _, ns := range nses {
		cmList, err := coreClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf(err.Error())
		}

		for _, cm := range cmList.Items {
			err := coreClient.CoreV1().ConfigMaps(cm.Namespace).Delete(cm.Name, &metav1.DeleteOptions{})
			if err != nil {
				t.Fatalf(err.Error())
			}
		}

		// wait for all items to be deleted
		for {
			cmList, err := coreClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf(err.Error())
			}

			if len(cmList.Items) == 0 {
				break
			}
		}
	}
}

func timeStep(description string, f func()) {
	t1 := time.Now()
	f()
	fmt.Printf("%s took %s\n", description, time.Now().Sub(t1))
}

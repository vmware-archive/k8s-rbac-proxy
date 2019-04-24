package e2e

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestMergedWatchResp(t *testing.T) {
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

	var watcher watch.Interface

	timeStep("config map watching (all ns)", func() {
		var err error

		watcher, err = coreClient.CoreV1().ConfigMaps("").Watch(metav1.ListOptions{})
		if err != nil {
			t.Fatalf(err.Error())
		}
	})

	eventsCh := make(chan watch.Event)
	cancelEventsCh := make(chan struct{})

	go func() {
		defer watcher.Stop()

		for {
			select {
			case e, ok := <-watcher.ResultChan():
				if !ok || e.Object == nil {
					t.Fatalf("expired before getting any events")
					return
				}

				eventsCh <- e

			case <-cancelEventsCh:
				return
			}
		}
	}()

	// Read off initial set of events after connecting
	for range nses {
		<-eventsCh
	}

	for _, ns := range nses {
		timeStep("make change and observe event", func() {
			cm, err := coreClient.CoreV1().ConfigMaps(ns).Get("test", metav1.GetOptions{})
			if err != nil {
				t.Fatalf(err.Error())
			}

			cm.Data = map[string]string{"changed": "true"}

			_, err = coreClient.CoreV1().ConfigMaps(ns).Update(cm)
			if err != nil {
				t.Fatalf(err.Error())
			}

			select {
			case event := <-eventsCh:
				if event.Object.(*corev1.ConfigMap).UID != cm.UID {
					t.Fatalf("Expected correct config map in the event")
				}

			case <-time.After(10 * time.Second):
				t.Fatalf("Expected event to arrive under 10s")
			}
		})
	}

	close(cancelEventsCh)
}

package main

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// TODO high QPS
	config.QPS = 100000
	config.Burst = 100000

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	t1 := time.Now()
	total := 0
	for i := 0; i < 10000; i++ {
		for _, ns := range []string{"rbac-test"} {
			pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			total += len(pods.Items)
		}
	}
	fmt.Printf("finished=%s, total=%d\n", time.Now().Sub(t1), total)
}

func main2() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("\n### list namespaces test\n")

	nsList, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d namespaces in cluster\n", len(nsList.Items))

	fmt.Printf("\n### list pods test\n")

	for _, ns := range []string{"rbac-test", "rbac-test2", ""} {
		fmt.Printf("listing in ns '%s'\n", ns)

		pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("There are %d pods\n", len(pods.Items))

		for _, pod := range pods.Items {
			fmt.Printf("pod: %s %s\n", pod.Namespace, pod.Name)
		}
	}

	fmt.Printf("\n### list role bindings test in rbac-test ns\n")

	rbList, err := clientset.RbacV1().RoleBindings("rbac-test").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d role bindings in ns\n", len(rbList.Items))

	watch(clientset)
}

func watch(clientset *kubernetes.Clientset) {
	fmt.Printf("\n### watch pods test across cluster\n")

	watcher, err := clientset.CoreV1().Pods("").Watch(metav1.ListOptions{})
	// watcher, err := clientset.CoreV1().ConfigMaps("").Watch(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	defer watcher.Stop()

	for {
		select {
		case e, ok := <-watcher.ResultChan():
			if !ok || e.Object == nil {
				// Watcher may expire, hence try to retry
				fmt.Printf("expired\n")
				return
			}

			fmt.Printf("object: %#v\n", e)
		}
	}
}

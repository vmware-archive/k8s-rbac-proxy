package proxy

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type TypeMetaResolver struct {
	coreClient kubernetes.Interface
}

func NewTypeMetaResolver(coreClient kubernetes.Interface) TypeMetaResolver {
	return TypeMetaResolver{coreClient}
}

func (r TypeMetaResolver) Resolve(pathMeta map[string]string) (metav1.TypeMeta, error) {
	serverResources, err := r.coreClient.Discovery().ServerResources()
	if err != nil {
		return metav1.TypeMeta{}, err
	}

	for _, resList := range serverResources {
		groupVersion, err := schema.ParseGroupVersion(resList.GroupVersion)
		if err != nil {
			return metav1.TypeMeta{}, err
		}

		for _, res := range resList.APIResources {
			group := groupVersion.Group
			if len(res.Group) > 0 {
				group = res.Group
			}

			version := groupVersion.Version
			if len(res.Version) > 0 {
				version = res.Version
			}

			if res.Name == pathMeta["resource"] && version == pathMeta["version"] && group == pathMeta["group"] {
				apiVersion := version
				if len(group) > 0 {
					apiVersion = group + "/" + apiVersion
				}

				// TODO better way to find list kind
				return metav1.TypeMeta{Kind: res.Kind + "List", APIVersion: apiVersion}, nil
			}
		}
	}

	return metav1.TypeMeta{}, fmt.Errorf("Expected to find metadata for '%#v'", pathMeta)
}

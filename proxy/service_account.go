package proxy

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceAccountFactory struct {
	coreClient kubernetes.Interface
}

func NewServiceAccountFactory(coreClient kubernetes.Interface) ServiceAccountFactory {
	return ServiceAccountFactory{coreClient}
}

// TODO cache service account namespaces
func (f ServiceAccountFactory) New(authHeaders []string) (ServiceAccount, error) {
	if len(authHeaders) == 0 {
		return ServiceAccount{}, fmt.Errorf("missing auth header")
	}

	tokenPieces := strings.SplitN(authHeaders[0], " ", 2) // remove bearer
	if len(tokenPieces) != 2 {
		return ServiceAccount{}, fmt.Errorf("removing token type")
	}

	claims := jwt.MapClaims{}

	// Proxy does not care if token is malicious because
	// it will continue to use this token to make data request
	// and API server will continue to verify actual access
	_, _, err := (&jwt.Parser{}).ParseUnverified(tokenPieces[1], &claims)
	if err != nil {
		return ServiceAccount{}, fmt.Errorf("parsing token: %s", err)
	}

	if claims["iss"] != "kubernetes/serviceaccount" {
		return ServiceAccount{}, fmt.Errorf("expected token to be issued for service account")
	}

	sa := ServiceAccount{
		coreClient: f.coreClient,
	}

	var ok bool

	sa.name, ok = claims["kubernetes.io/serviceaccount/service-account.name"].(string)
	if !ok {
		return ServiceAccount{}, fmt.Errorf("expected token to have service account name")
	}

	sa.nsName, ok = claims["kubernetes.io/serviceaccount/namespace"].(string)
	if !ok {
		return ServiceAccount{}, fmt.Errorf("expected token to have service account namespace")
	}

	return sa, nil
}

type ServiceAccount struct {
	name       string
	nsName     string
	coreClient kubernetes.Interface
}

func (s ServiceAccount) Namespaces() ([]string, error) {
	bindingsList, err := s.coreClient.RbacV1().RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("fetching all role bindings: %s", err)
	}

	var nsNames []string

	for _, binding := range bindingsList.Items {
		for _, subject := range binding.Subjects {
			if subject.Name == s.name && subject.Namespace == s.nsName && subject.Kind == "ServiceAccount" {
				nsNames = append(nsNames, binding.Namespace)
			}
		}
	}

	return nsNames, nil
}

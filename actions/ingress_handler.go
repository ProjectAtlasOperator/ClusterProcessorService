package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type IngressInformation struct {
	Host string `json:"host"`
}

func IngressHander(c buffalo.Context) error {
	namespace := c.Param("namespace")
	fmt.Println("=> Using namespace: " + namespace)
	if len(namespace) == 0 {
		fmt.Println("=> Url Param 'namespace' is missing. Using 'default' namespace")
		namespace = "default"
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ingresses, err := clientset.ExtensionsV1beta1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	ingressInformations := []IngressInformation{}

	for _, ingress := range ingresses.Items {
		ingressInformation := IngressInformation{}
		for _, rule := range ingress.Spec.Rules {
			ingressInformation.Host = rule.Host
		}
		ingressInformations = append(ingressInformations, ingressInformation)
	}

	return c.Render(http.StatusOK, r.JSON(ingressInformations))
}

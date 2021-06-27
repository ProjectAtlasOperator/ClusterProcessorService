package actions

import (
	"context"
	"net/http"

	"github.com/gobuffalo/buffalo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Namespace [10]struct {
	Name string `json:"name"`
}

func NamespaceHander(c buffalo.Context) error {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	var ns Namespace

	for i := 0; i < len(namespaces.Items); i++ {
		ns[i].Name = namespaces.Items[i].Name
	}

	return c.Render(http.StatusOK, r.JSON(ns))

}

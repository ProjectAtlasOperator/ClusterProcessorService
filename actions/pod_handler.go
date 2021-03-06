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

// PodInfoHander is a default handler to serve up a home page.
func PodInfoHander(c buffalo.Context) error {
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

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return c.Render(http.StatusOK, r.JSON(pods.Items))
}

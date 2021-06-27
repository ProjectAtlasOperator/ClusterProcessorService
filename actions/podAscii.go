package actions

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodInfoHander is a default handler to serve up a home page.
func PodAsciiHander(c buffalo.Context) error {
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

	var asciiDocString = strings.Builder{}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(pods.Items); i++ {
		for _, pod := range pods.Items {
			asciiDocString.WriteString("|===\nPOD | INFO\n")
			asciiDocString.WriteString("\n\n|PodName\n|")
			asciiDocString.WriteString(pods.Items[i].Name)

			asciiDocString.WriteString("\n\n|HostIP\n|")
			asciiDocString.WriteString(pods.Items[i].Status.HostIP)

			asciiDocString.WriteString("\n\n|PodIP\n|")
			asciiDocString.WriteString(pods.Items[i].Status.PodIP)

			asciiDocString.WriteString("\n\n|StarTime\n|")
			asciiDocString.WriteString(pods.Items[i].Status.StartTime.Time.String())

			asciiDocString.WriteString("\n\n|===\n\n")
			if len(pods.Items) > 0 {
				for _, container := range pod.Spec.Containers {
					envVariables := container.Env
					if len(envVariables) > 0 {
						asciiDocString.WriteString("|===\n|ENV |VALUE\n\n|")
						for _, spec := range envVariables {
							asciiDocString.WriteString(spec.Name)
							asciiDocString.WriteString("\n|")
							asciiDocString.WriteString(spec.Value)
							//asciiDocString.WriteString(spec.ValueFrom.SecretKeyRef.Name)
							asciiDocString.WriteString("\n|")
						}
						asciiDocString.WriteString("\n|===\n\n")
					}
				}
			}
		}
	}
	return c.Render(http.StatusCreated, r.Download(c, "ascii.adoc", strings.NewReader(asciiDocString.String())))
}

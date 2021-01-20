package actions

import (
	"context"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"
	"strings"
)

// PodInfoHander is a default handler to serve up a home page.
func PodAsciiHander(c buffalo.Context) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientsetMetrics, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var asciiDocString = strings.Builder{}
	//podInformation := &PodInformation{}
	//var _podArray [4]byte
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("project-atlas-system").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		nodes, err := clientsetMetrics.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}


		for i := 0; i < len(pods.Items); i++ {
			for _, pod := range pods.Items {
				asciiDocString.WriteString("|===\nPOD | INFO\n" )
				asciiDocString.WriteString("\n\n|PodName\n|" )
				asciiDocString.WriteString(pods.Items[i].Name)

				asciiDocString.WriteString("\n\n|Namespace\n|" )
				asciiDocString.WriteString(pods.Items[i].Namespace)

				asciiDocString.WriteString("\n\n|HostIP\n|" )
				asciiDocString.WriteString(pods.Items[i].Status.HostIP)

				asciiDocString.WriteString("\n\n|PodIP\n|" )
				asciiDocString.WriteString(pods.Items[i].Status.PodIP)

				asciiDocString.WriteString("\n\n|StarTime\n|" )
				asciiDocString.WriteString(pods.Items[i].Status.StartTime.Time.String())

				asciiDocString.WriteString("\n\n|===\n\n" )
				//podInformation[i].MDBPort = int(pods.Items[i].Spec.Containers[i].Ports[i].ContainerPort)
				//podInformation[i].VolumeName = pods.Items[i].Spec.Containers[i].VolumeMounts[i].Name
				//podInformation[i].VolumeMount = pods.Items[i].Spec.Containers[i].VolumeMounts[i].MountPath
				//podInformation[i].CPUUsage = podsMetricList.Items[i].Containers[i].Usage.Cpu().String()
				//podInformation[i].MemoryUsage = podsMetricList.Items[i].Containers[i].Usage.Memory().String()
				//podInformation[i].ConfigMapName = configMap.Items[i].Name
				//podInformation[i].NodeName = nodes.Items[i].Name

				podStatus := pod.Spec.Containers
				for _, spec := range podStatus {
					asciiDocString.WriteString("|===\n|ENV |VALUE\n\n|" )
					asciiDocString.WriteString(spec.Env[0].Name)
					asciiDocString.WriteString("\n|" )
					asciiDocString.WriteString(spec.Env[0].Value)
					asciiDocString.WriteString("\n\n|===\n\n" )
				}
			}
		}

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		c.Set("nodeName", nodes.Items[0].Name)
		c.Set("nodeMemory", nodes.Items[0].Usage.Memory())

		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("project-atlas-system").Get(context.TODO(), "example-xxxxx", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found example-xxxxx pod in default namespace\n")
		}
		break
	}

	return c.Render(http.StatusCreated,  r.Download(c,"ascii.adoc", strings.NewReader(asciiDocString.String())))

}


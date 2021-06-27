package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodInformations struct {
	PodInformations [4]PodInformation `json:"podInformations"`
}

type PodInformation struct {
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
	StartTime string `json:"startTime"`

	VolumeName  string `json:"volumeName"`
	VolumeMount string `json:"volumeMount"`

	CPUPodName  string `json:"cpuPodName"`
	CPUUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`

	ImageName string `json:"imageName"`
	MountPath string `json:"mountPath"`

	ConfigMapName string `json:"configMapName"`
}

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
	clientsetMetrics, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var podInformations PodInformations

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	podsMetricList, err := clientsetMetrics.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(pods.Items); i++ {
		for _, pod := range pods.Items {
			podInformations.PodInformations[i].PodName = pods.Items[i].Name

			podStatus := pod.Spec.Containers
			for _, spec := range podStatus {
				c.Set("image", spec.Image)
				c.Set("imageName", spec.Name)
				volume := spec.VolumeMounts
				for _, volume := range volume {
					podInformations.PodInformations[i].VolumeName = volume.Name
					podInformations.PodInformations[i].VolumeMount = volume.MountPath
					c.Set("mountPath", volume.MountPath)
					c.Set("volumeName", volume.Name)
				}
			}
		}

		for _, podMetric := range podsMetricList.Items {
			podContainer := podMetric.Containers
			for _, container := range podContainer {
				podInformations.PodInformations[i].CPUPodName = container.Name
				podInformations.PodInformations[i].CPUUsage = container.Usage.Cpu().String()
				podInformations.PodInformations[i].MemoryUsage = container.Usage.Memory().String()

				NAME := container.Name
				CPU := container.Usage.Cpu().AsDec()
				MEMORY := container.Usage.Memory()
				c.Set("NAME", NAME)
				c.Set("CPU", CPU)
				c.Set("MEMORY", MEMORY)
			}
		}

		for _, configmap := range configMap.Items {
			podInformations.PodInformations[i].ConfigMapName = configmap.Name
			c.Set("cfm", configmap.Name)
			c.Set("cfm_data", configmap.Data)
		}
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	return c.Render(http.StatusOK, r.JSON(podInformations))
}

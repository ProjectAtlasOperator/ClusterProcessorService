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

	podInformations := []PodInformation{}

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
		podInformation := PodInformation{}

		for _, pod := range pods.Items {
			podInformation.PodName = pods.Items[i].Name

			podStatus := pod.Spec.Containers
			for _, spec := range podStatus {
				c.Set("image", spec.Image)
				c.Set("imageName", spec.Name)
				volume := spec.VolumeMounts
				for _, volume := range volume {
					podInformation.VolumeName = volume.Name
					podInformation.VolumeMount = volume.MountPath
					c.Set("mountPath", volume.MountPath)
					c.Set("volumeName", volume.Name)
				}
			}
		}

		for _, podMetric := range podsMetricList.Items {
			podContainer := podMetric.Containers
			for _, container := range podContainer {
				podInformation.CPUPodName = container.Name
				podInformation.CPUUsage = container.Usage.Cpu().String()
				podInformation.MemoryUsage = container.Usage.Memory().String()

				NAME := container.Name
				CPU := container.Usage.Cpu().AsDec()
				MEMORY := container.Usage.Memory()
				c.Set("NAME", NAME)
				c.Set("CPU", CPU)
				c.Set("MEMORY", MEMORY)
			}
		}

		for _, configmap := range configMap.Items {
			podInformation.ConfigMapName = configmap.Name
			c.Set("cfm", configmap.Name)
			c.Set("cfm_data", configmap.Data)
		}
		podInformations = append(podInformations, podInformation)
	}
	return c.Render(http.StatusOK, r.JSON(podInformations))
}

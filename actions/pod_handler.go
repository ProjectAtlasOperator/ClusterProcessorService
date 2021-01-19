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
)

type PodInformations struct {
	PodInformations [4]PodInformation `json:"podInformations"`
}

type PodInformation struct {
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
	HostIP    string `json:"hostIp"`
	PodIP     string `json:"podIP"`
	StartTime string `json:"startTime"`

	//VolumePodName string `json:"volumePodName"`
	VolumeName  string `json:"volumeName"`
	VolumeMount string `json:"volumeMount"`

	CPUPodName  string `json:"cpuPodName"`
	CPUUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`

	//ImageName string `json:"imageName"`
	//MountPath string `json:"mountPath"`

	ConfigMapName string `json:"configMapName"`

	NodeName     string `json:"NodeName"`
	NodeMemory   string `json:"nodeMemory"`
	MDBPort      int    `json:"mdbPort"`
	MExpressPort int    `json:"mExpressPort"`
	CPSPort      int    `json:"cpsPort"`
}

// PodInfoHander is a default handler to serve up a home page.
func PodInfoHander(c buffalo.Context) error {
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
	//var _podArray [4]byte
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("project-atlas-system").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		configMap, err := clientset.CoreV1().ConfigMaps("project-atlas-system").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		podsMetricList, err := clientsetMetrics.MetricsV1beta1().PodMetricses("project-atlas-system").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		nodes, err := clientsetMetrics.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for i := 0; i < len(pods.Items); i++ {
			for _, pod := range pods.Items {
				podInformations.PodInformations[i].PodName = pods.Items[i].Name
				podInformations.PodInformations[i].Namespace = pods.Items[i].Namespace
				podInformations.PodInformations[i].HostIP = pods.Items[i].Status.HostIP
				podInformations.PodInformations[i].PodIP = pods.Items[i].Status.PodIP
				podInformations.PodInformations[i].StartTime = pods.Items[i].Status.StartTime.Time.String()
				//podInformation[i].MDBPort = int(pods.Items[i].Spec.Containers[i].Ports[i].ContainerPort)
				//podInformation[i].VolumeName = pods.Items[i].Spec.Containers[i].VolumeMounts[i].Name
				//podInformation[i].VolumeMount = pods.Items[i].Spec.Containers[i].VolumeMounts[i].MountPath
				//podInformation[i].CPUUsage = podsMetricList.Items[i].Containers[i].Usage.Cpu().String()
				//podInformation[i].MemoryUsage = podsMetricList.Items[i].Containers[i].Usage.Memory().String()
				//podInformation[i].ConfigMapName = configMap.Items[i].Name
				//podInformation[i].NodeName = nodes.Items[i].Name

				podStatus := pod.Spec.Containers
				for _, spec := range podStatus {
					c.Set("image", spec.Image)
					c.Set("imageName", spec.Name)
					volume := spec.VolumeMounts
					for _, port := range spec.Ports {
						podInformations.PodInformations[i].MDBPort = int(port.ContainerPort)
					}
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

	return c.Render(http.StatusOK, r.JSON(podInformations))
}

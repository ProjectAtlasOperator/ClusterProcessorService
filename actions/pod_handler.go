package actions

import (
	"context"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"
)

type PodInformation struct {
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
	HostIP    string `json:"hostIp"`
	PodIP     string `json:"podIp"`
	StartTime string `json:"startTime"`

	volume struct {
		PodName     string `json:"podName"`
		VolumeName  string `json:"volumeName"`
		VolumeMount string `json:"volumeMount"`
	}
	pod struct {
		Name      string `json:"name"`
		Namespace string `json:"nameSpace"`
		PodIP     string `json:"podIp"`
		HostIP    string `json:"hostIp"`
		StartTime string `json:"startTime"`
		Port 	  string `json:"port"`
	}
	cpuMem struct {
		PodName     string `json:"podName"`
		CPUUsage    string `json:"cpuUsage"`
		MemoryUsage string `json:"memoryUsage"`
	}
	image struct {
		ImageName  string `json:"imageName"`
		MountPath  string `json:"mountPath"`
		VolumeName string `json:"volumeName"`
	}
	configMap struct {
		Name string `json:"name"`
	}
	node struct {
		Name   string `json:"name"`
		Memory string `json:"memory"`
	}
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

	podInformation := &PodInformation{}

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

		//getPodInfo(pods, p, &ctx, &c)
		for _, pod := range pods.Items {
			podInformation.volume.PodName = pod.Name
			podInformation.pod.Name = pod.Name
			podInformation.pod.Namespace = pod.Namespace
			podInformation.pod.HostIP = pod.Status.HostIP
			podInformation.pod.PodIP = pod.Status.PodIP
			podInformation.pod.StartTime = pod.Status.StartTime.Time.String()
			//podInformation.pod.Port = pod.spec.container

			podStatus := pod.Spec.Containers
			for _, spec := range podStatus {
				c.Set("image", spec.Image)
				c.Set("imageName", spec.Name)
				volume := spec.VolumeMounts
				for _, volume := range volume {
					podInformation.volume.VolumeName = volume.Name
					podInformation.volume.VolumeMount = volume.MountPath
					c.Set("mountPath", volume.MountPath)
					c.Set("volumeName", volume.Name)
				}
			}
		}
		for _, podMetric := range podsMetricList.Items {
			podContainer := podMetric.Containers
			for _, container := range podContainer {
				podInformation.cpuMem.PodName = container.Name
				podInformation.cpuMem.CPUUsage = container.Usage.Cpu().String()
				podInformation.cpuMem.MemoryUsage = container.Usage.Memory().String()

				NAME := container.Name
				CPU := container.Usage.Cpu().AsDec()
				MEMORY := container.Usage.Memory()
				c.Set("NAME", NAME)
				c.Set("CPU", CPU)
				c.Set("MEMORY", MEMORY)
			}
		}

		for _, configmap := range configMap.Items {
			podInformation.configMap.Name = configmap.Name
			c.Set("cfm", configmap.Name)
			c.Set("cfm_data", configmap.Data)
		}

		for _, node := range nodes.Items {
			podInformation.node.Name = node.Name
			podInformation.node.Memory = node.Usage.Memory().String()
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		c.Set("pods", pods)
		c.Set("nodeName", nodes.Items[0].Name)
		c.Set("nodeMemory", nodes.Items[0].Usage.Memory())

		//return c.Render(http.StatusOK, r.HTML("index.html"))
		//if err := c.Bind(p.pod); err != nil {
		//	return err
		//} // was
		if err := c.Bind(podInformation.pod); err != nil {
			return err
		}

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

	fmt.Println(podInformation)

	e, err := json.Marshal(podInformation)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(e)

	//return c.Render(http.StatusOK, r.HTML("pod-handler.html"))
	return c.Render(http.StatusOK, r.JSON(e))
}

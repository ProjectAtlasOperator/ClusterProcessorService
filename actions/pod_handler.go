package actions

import (
	"context"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"net/http"
	"time"
)

type Pod struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty"`
	Namespace string             `json:"namespace,omitempty"`
	PodIP     string             `json:"podIp,omitempty"`
	HostIP    string             `json:"hostIp,omitempty"`
	StartTime time.Time          `json:"startTime,omitempty"`
	//cpu *inf.Dec        `json:"cpu,omitempty"`
	//memory *inf.Dec          `json:"memory,omitempty"`
}
type Node struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name   string             `json:"name,omitempty"`
	Memory string             `json:"memory,omitempty"`
	//cpu *inf.Dec        `json:"cpu,omitempty"`
}
type ConfigMap struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name,omitempty"`
	//cpu *inf.Dec        `json:"cpu,omitempty"`
}
type Volume struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PodName     string             `json:"podname,omitempty"`
	VolumeName  string             `json:"volumename,omitempty"`
	VolumeMount string             `json:"volumemount,omitempty"`
	//cpu *inf.Dec        `json:"cpu,omitempty"`
}
type CpuMem struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PodName     string             `json:"podname,omitempty"`
	CPUUsage    string             `json:"cpuusage,omitempty"`
	MemoryUsage string             `json:"memoryusage,omitempty"`
	//cpu *inf.Dec        `json:"cpu,omitempty"`
}

// PodInfoHander is a default handler to serve up
// a home page.
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

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://username:password@10.97.103.216:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//Connection to MDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

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

		p := &Pod{}
		v := &Volume{}
		//getPodInfo(pods, p, &ctx, &c)
		for _, pod := range pods.Items {
			v.PodName = pod.Name
			p.Name = pod.Name
			p.Namespace = pod.Namespace
			p.HostIP = pod.Status.HostIP
			p.PodIP = pod.Status.PodIP
			p.StartTime = pod.Status.StartTime.Time

			addPodToDatabase(&ctx, p, client)

			podStatus := pod.Spec.Containers
			for _, spec := range podStatus {
				c.Set("image", spec.Image)
				c.Set("imageName", spec.Name)
				volume := spec.VolumeMounts
				for _, volume := range volume {
					v.VolumeName = volume.Name
					v.VolumeMount = volume.MountPath
					c.Set("mountPath", volume.MountPath)
					c.Set("volumeName", volume.Name)
				}
			}
			addVolumeToDatabase(&ctx, v, client)
		}
		cpumem := &CpuMem{}
		for _, podMetric := range podsMetricList.Items {
			podContainer := podMetric.Containers
			for _, container := range podContainer {
				cpumem.PodName = container.Name
				cpumem.CPUUsage = container.Usage.Cpu().String()
				cpumem.MemoryUsage = container.Usage.Memory().String()

				NAME := container.Name
				CPU := container.Usage.Cpu().AsDec()
				MEMORY := container.Usage.Memory()
				//p.cpu = CPU
				//p.memory = MEMORY
				c.Set("NAME", NAME)
				c.Set("CPU", CPU)
				c.Set("MEMORY", MEMORY)
			}
			addCpuMemToDatabase(&ctx, cpumem, client)

		}

		confMap := &ConfigMap{}
		for _, configmap := range configMap.Items {
			confMap.Name = configmap.Name
			c.Set("cfm", configmap.Name)
			c.Set("cfm_data", configmap.Data)
			addConfMapToDatabase(&ctx, confMap, client)

		}

		n := &Node{}
		for _, node := range nodes.Items {
			n.Name = node.Name
			n.Memory = node.Usage.Memory().String()
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		c.Set("pods", pods)
		c.Set("nodeName", nodes.Items[0].Name)
		c.Set("nodeMemory", nodes.Items[0].Usage.Memory())
		addNodeToDatabase(&ctx, n, client)

		//return c.Render(http.StatusOK, r.HTML("index.html"))
		if err := c.Bind(p); err != nil {
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
		//time.Sleep(2 * time.Second)

	}

	return c.Render(http.StatusOK, r.HTML("pod-handler.html"))
}

func addNodeToDatabase(ctx *context.Context, n *Node, client *mongo.Client) {
	collection := client.Database("project-atlas").Collection("node")
	_, err := collection.InsertOne(*ctx, n)
	if err != nil {
		log.Fatal(err)
	}
}
func addVolumeToDatabase(ctx *context.Context, v *Volume, client *mongo.Client) {
	collection := client.Database("project-atlas").Collection("volume")
	_, err := collection.InsertOne(*ctx, v)
	if err != nil {
		log.Fatal(err)
	}
}
func addCpuMemToDatabase(ctx *context.Context, cpumem *CpuMem, client *mongo.Client) {
	collection := client.Database("project-atlas").Collection("cpu-memory")
	_, err := collection.InsertOne(*ctx, cpumem)
	if err != nil {
		log.Fatal(err)
	}
}
func addPodToDatabase(ctx *context.Context, p *Pod, client *mongo.Client) {
	collection := client.Database("project-atlas").Collection("pod")
	_, err := collection.InsertOne(*ctx, p)
	if err != nil {
		log.Fatal(err)
	}
}
func addConfMapToDatabase(ctx *context.Context, confMap *ConfigMap, client *mongo.Client) {
	collection := client.Database("project-atlas").Collection("config-map")
	_, err := collection.InsertOne(*ctx, confMap)
	if err != nil {
		log.Fatal(err)
	}
}

func getPodInfo(pods *v1.PodList, p *Pod, ctx *context.Context, c buffalo.Context) {
	for _, pod := range pods.Items {
		p.Name = pod.Name
		p.Namespace = pod.Namespace
		p.HostIP = pod.Status.HostIP
		p.PodIP = pod.Status.PodIP
		p.StartTime = pod.Status.StartTime.Time

		collection := client.Database("project-atlas").Collection("pod")
		_, err := collection.InsertOne(*ctx, p)
		if err != nil {
			log.Fatal(err)
		}

		podStatus := pod.Spec.Containers
		for _, spec := range podStatus {
			c.Set("image", spec.Image)
			c.Set("imageName", spec.Name)
			volume := spec.VolumeMounts
			for _, volume := range volume {
				c.Set("mountPath", volume.MountPath)
				c.Set("volumeName", volume.Name)
			}
		}
	}
}

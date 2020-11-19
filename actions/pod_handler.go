package actions

import (
	"context"
	"fmt"
	"github.com/gobuffalo/buffalo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		for _, pod := range pods.Items {
			p.Name = pod.Name
			p.Namespace = pod.Namespace
			p.HostIP = pod.Status.HostIP
			p.PodIP = pod.Status.PodIP
			p.StartTime = pod.Status.StartTime.Time

			collection := client.Database("local").Collection("pod")
			_, err = collection.InsertOne(ctx, p)
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

		for _, podMetric := range podsMetricList.Items {
			podContainer := podMetric.Containers
			for _, container := range podContainer {
				CPU := container.Usage.Cpu().AsDec()
				MEMORY := container.Usage.Memory().AsDec()
				c.Set("CPU", CPU)
				c.Set("MEMORY", MEMORY)
			}
		}

		for _, configmap := range configMap.Items {
			c.Set("cfm", configmap.Name)
			c.Set("cfm_data", configmap.Data)
		}

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		c.Set("pods", pods)
		c.Set("nodeName", nodes.Items[0].Name)
		c.Set("nodeMemory", nodes.Items[0].Usage.Memory())

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

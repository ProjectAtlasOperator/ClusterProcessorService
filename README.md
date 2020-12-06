# Welcome to Buffalo!

This project was scaffolded via command:

	buffalo new cluster-processor-service --skip-yarn --skip-webpack --skip-pop

## Starting the Application

Buffalo ships with a command that will watch your application and automatically rebuild the Go binary and any assets for you. To do that run the "buffalo dev" command:

	buffalo dev

If you point your browser to [http://127.0.0.1:3000](http://127.0.0.1:3000) you should see a "Welcome to Buffalo!" page.

**Congratulations!** You now have your Buffalo application up and running.

## Docker build

    docker build . --file Dockerfile --tag marianferenc/project_atlas_cluster_processor_service:latest
	
## Kubernetes deployment

	helm upgrade --install cluster-processor-service charts/cluster_processor_service --namespace project-atlas-system
	
## Kubernetes undeployment

    helm del cluster-processor-service --namespace project-atlas-system
    
## Forward deployed application to localhost

    kubectl port-forward svc/cluster-processor-service 3000:3000 --namespace project-atlas-system
    
## Check logs on kubernetes
First get name of pod instance via command:

    kubectl get pods --namespace project-atlas-system

Sample output:

```
NAME                                        READY   STATUS              RESTARTS   AGE
cluster-processor-service-7dd75b65c-hzjvv   1/1     Running               0      4m21s
```

Now you can get logs from this pod via command:

    kubectl logs -f cluster-processor-service-7dd75b65c-hzjvv
    
BONUS (one liner for people with linux terminal):
 
    kubectl get pods -o custom-columns=POD:.metadata.name --no-headers | grep cluster-processor-service | xargs kubectl logs -f

##Start Metrics Server

Download metrics server from GitHub

    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

Change policies of metrics server deployment
    
    kubectl -n kube-system edit deployments.apps metrics-server

Add this under spec: containers: - args:

            command:
            - /metrics-server
            - --kubelet-insecure-tls
            - --kubelet-preferred-address-types=InternalIP
            
Use command kubectl top pods or nodes to access CPU and MEMORY usage

    kubectl top pods
    NAME                                         CPU(cores)   MEMORY(bytes)
    cluster-processor-service-7944949d67-h8k6w   1m           14Mi

Use command kubectl top pods or nodes to access CPU and MEMORY usage (osx)

    kubectl top pods --namespace project-atlas-system
    NAME                                         CPU(cores)   MEMORY(bytes)
    cluster-processor-service-7944949d67-h8k6w   1m           14Mi    

## Start configmap

Enter directory yaml_files and start configmap with
    
    kubectl apply -f configmap.yaml --namespace project-atlas-system
    
## Start configmap

Create namespace
    
    kubectl create namespace project-atlas-system   

## What Next?

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

Good luck!

[Powered by Buffalo](http://gobuffalo.io)

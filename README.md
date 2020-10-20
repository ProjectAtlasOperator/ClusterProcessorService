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

	helm install cluster-processor-service charts/cluster_processor_service
	
## What Next?

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

Good luck!

[Powered by Buffalo](http://gobuffalo.io)

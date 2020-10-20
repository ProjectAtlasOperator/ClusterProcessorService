module cluster_processor_service

go 1.13

require (
	github.com/gobuffalo/buffalo v0.16.15
	github.com/gobuffalo/envy v1.9.0
	github.com/gobuffalo/mw-csrf v1.0.0
	github.com/gobuffalo/mw-forcessl v0.0.0-20200131175327-94b2bd771862
	github.com/gobuffalo/mw-i18n v1.1.0
	github.com/gobuffalo/mw-paramlogger v1.0.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobuffalo/suite v2.8.2+incompatible
	github.com/unrolled/secure v1.0.8
	k8s.io/api v0.19.2
	//k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.0.0-20201007120136-5682372f3538
	k8s.io/api/apiserverinternal/v1alpha1 v0.0.0-20201005155906-2c3c141c931c
)

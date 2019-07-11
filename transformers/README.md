# transformers
Reseach: Queue-oriented, Model-free Architecture for Various Application Lifecycle Management

# Prerequisite

Ensure you have configured GOROOT and GOPATH

- IDE: GoLand 
- Go: >= 1.10.3
- go get -u -v k8s.io/client-go
- go get -u -v k8s.io/kubernetes

# Deploy Steps
#### Step 1: Add code to kube-scheduler source code
Find `$GOPATH/k8s.io/kubernetes/pkg/scheduler/scheduler.go`, namely original `scheduler.go`.
##### Method1:
Copy all the code snippets in file `<RepoPath>/diff.go` to the end of `scheduler.go`, resolve all the import and package problems, then find `func (sched *Scheduler) Run()` in original `scheduler.go`(around line 281), change it as shown below.

```go
func (sched *Scheduler) Run() {
	if !sched.config.WaitForCacheSync() {
		return
	}
	go wait.Until(sched.scheduleOne, 0, sched.config.StopEverything)
	go transfomers()
}
```

find `func(sched *Scheduler) scheduleOne()` in original `scheduler.go`, insert several lines of code behind `suggestedHost, err := sched.schedule(pod)`
(around line 540). After the insertion, `func(sched *Scheduler) scheduleOne()` should look like this:

```go
func (sched *Scheduler) scheduleOne() {
	[...]
	// Synchronously attempt to find a fit for the pod.
	start := time.Now()
	suggestedHost, err := sched.schedule(pod)

	// Inserted
	if pod.GetAnnotations()["proxy"] != "" {
		Hack(suggestedHost, pod)
		return
	}

	if err != nil {
		// schedule() may have failed because the pod would not fit on any host, so we try to
		// preempt, with the expectation that the next time the pod is tried for scheduling it
		// will fit due to the preemption. It is also possible that a different pod will schedule
		// into the resources that were preempted, but this is harmless.
		if fitError, ok := err.(*core.FitError); ok {
		[...]
	   
```

##### Method2:
Just replace the original `scheduler.go` with `<RepoPath>/replacement/scheduler.go`

#### Step 2: Rebuild scheduler
Move to `$GOPATH/k8s.io/kubernetes/cmd/kube-scheduler/`, and build with command `env GOOS=linux GOARCH=amd64 go build scheduler.go`, now you get a executable binary file `scheduler`.
#### Step 3: Pack as a image
Build the image with `<RepoPath>/build/Dockerfile`, before that you need to move executable scheduler binary file to /build path, then build image with command.

```
docker build -t custom-scheduler:1.0 .
```
#### Step 4: Deploy
Created a Deployment configuration file and ran it in an existing Kubernetes cluster, using the Deployment resource rather than creating a Pod resource directly because the Deployment resource can better handle the case of a scheduler running node failure. We offered a Deployment configuration example, saved as the `custom-scheduler.yaml` in `<RepoPath>/yaml/`.

Then Create Deployment resource in Kubernetes cluster

```
kubectl create -f custom-scheduler.yaml
```

# Try it out

Once you have deployed `custom-scheduler` following the steps above, you can now try it out.
#### Step1: Create CRD
Create the CustomResourceDefinition you need. For example, a typical CRD with kind `myapp`.

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: myapps.app.example.com
spec:
  group: app.example.com
  names:
    kind: MyApp
    listKind: MyAppList
    plural: myapps
    singular: myapp
  scope: Namespaced
  version: v1alpha1
```

Then apply it to Kubernetes cluster.

```
kubectl apply -f myapp_crd.yaml
```
#### Step2: Create ConfigMap
Create a ConfigMap named `custom-resources`, list all the CustomResources kinds you need to watch in `data` field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-resources
  namespace: default
data:
  resources.1: myapps
```

Then apply it to Kubernetes cluster.

```
kubectl apply -f configmap.yaml
```

#### Step3: Create CustomResource Objects
Create CustomResource objects with `test-app1.yaml`:
```yaml
apiVersion: app.example.com/v1alpha1
kind: MyApp
metadata:
  name: test-app1
spec:
  requests:
    cpu: 500m
    memory: 500Mi
  input: "Hello World!"
  fileURL: "https://github.com/operator-framework/operator-sdk/blob/master/release.sh"
```
Apply `test-app1.yaml` to cluster.
```
kubectl apply -f test-app1.yaml
```
Then, when you use `kubectl describe myapp test-app1` to get test-app1's infomation, you can find that `Node Name` in `spec` field has been updated to the name of the node where test-app1 should be scheduled:


```Name:         test-app1
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"app.example.com/v1alpha1","kind":"MyApp","metadata":{"annotations":{},"name":"test-app1","namespace":"default"},"spec":{"fi...
API Version:  app.example.com/v1alpha1
Kind:         MyApp
Metadata:
  Creation Timestamp:  2019-03-21T07:55:39Z
  Generation:          1
  Resource Version:    190811
  Self Link:           /apis/app.example.com/v1alpha1/namespaces/default/myapps/test-app1
  UID:                 b7dc0154-4bae-11e9-b60a-00163e034c99
Spec:
  File URL:  https://github.com/operator-framework/operator-sdk/blob/master/release.sh
  Input:     Hello World!
  Node Name: master
  Requests:
    Cpu:     500m
    Memory:  250Mi
Events:      <none>
```
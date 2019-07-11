const(
	file = "/etc/kubernetes/admin.conf"
)

func Hack(suggestedHost string, pod *v1.Pod) {

	// Hack
	dynamicClient, err := NewDynamicClient()
	if err != nil {
		klog.Errorf("new dynamic client error")
	}

	//client, err := NewClient()
	if err != nil {
		klog.Errorf("new dynamic client error")
	}

	name := pod.GetAnnotations()["kind"]
	var gvr = schema.GroupVersionResource {
		Group:    "app.example.com",
		Version:  "v1alpha1",
		Resource: name,
	}
	// update CR's spec.nodeName
	// for more information, refer to JSON PATCH method standard
	if suggestedHost != "" {
		patchString := fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, suggestedHost)
		dynamicClient.Resource(gvr).
			Namespace("default").
			Patch(pod.GetName(), types.MergePatchType, []byte(patchString), mtv1.UpdateOptions{})
		got, _ := dynamicClient.Resource(gvr).
			Namespace("default").
			Get(pod.GetName(),metav1.GetOptions{})

		fmt.Println(got.Object)
	}

	// the pod no more needed, delete it
	//err = client.CoreV1().
	//	Pods("default").
	//	Delete(pod.GetName(), &metav1.DeleteOptions{})
	//if err != nil {
	//	fmt.Println(err)
	//}
}


func transformer() {
	for ; ; {
		resourceNames := GetCustomResourceNamesFromConfigMap()
		CreateNamespaces(resourceNames)
		for _, resourceName := range resourceNames {

			go ScheduleCustomResources(resourceName)
			time.Sleep(time.Second * 3)
		}
	}
}

func NewDynamicClient() (dynamic.Interface, error) {
	bytes, _ := ioutil.ReadFile(file)
	config, _ := clientcmd.NewClientConfigFromBytes(bytes)
	clientConfig, _ := config.ClientConfig()
	return dynamic.NewForConfig(clientConfig)
}

func NewClient() (*kubernetes.Clientset, error) {
	bytes, _ := ioutil.ReadFile(file)
	config, _ := clientcmd.NewClientConfigFromBytes(bytes)
	clientConfig, _ := config.ClientConfig()
	return kubernetes.NewForConfig(clientConfig)
}


func GetCustomResourceNamesFromConfigMap() (names []string) {
	client, err := NewClient()

	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}

	got, _ := client.
		CoreV1().
		ConfigMaps("default").
		Get("custom-resources", metav1.GetOptions{})

	for _, name := range got.Data {
		names = append(names, name)
	}

	return
}

func CreateNamespaces(names []string) {
	client, err := NewClient()
	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}

	for _, name := range names {
		namespaceName := name + "-ns"
		got, _ := client.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if(got == nil) {
			var ns = v1.Namespace{}
			ns.APIVersion = "v1"
			ns.Kind = "Namespace"
			ns.Name = namespaceName
			_, err := client.CoreV1().Namespaces().Create(&ns)
			if err != nil {
				klog.Errorf("create namespace error: %s", err)
				panic(err)
			}
		}
	}
}
func ScheduleCustomResources(name string) {

	dynamicClient, err := NewDynamicClient()

	if err != nil {
		klog.Errorf("create dynamic error: %s", err)
		panic(err)
	}

	client, err := NewClient()
	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}


	// TODO: Custom Group & Version
	var gvr = schema.GroupVersionResource{
		Group:    "app.example.com",
		Version:  "v1alpha1",
		Resource: name,
	}



	got, _ := dynamicClient.Resource(gvr).
		Namespace("default").
		List(metav1.ListOptions{})



	for _, item := range got.Items {

		nodeName := item.Object["spec"].(map[string]interface{})["nodeName"]
		fmt.Println(nodeName)

		// if nodeName == nil, it's not scheduled yet
		if nodeName == nil {

			pod := Transform(item)
			pod.Namespace = name + "-ns"
			fmt.Println(name + "-ns")
			got, err := client.CoreV1().
				Pods(name + "-ns").
				Create(&pod)
			fmt.Println(got, err)

		}

	}
}


// Transform (item Unstructured) to (pod v1.Pod)
// Need to improve
// TODO: Consider the resource requests and limitations

func Transform(un unstructured.Unstructured) v1.Pod {
	var pod v1.Pod

	jsonBytes, err := json.Marshal(un.Object)
	if err != nil {
		klog.Errorf("json marshal error")
		panic(err)
	}
	json.Unmarshal(jsonBytes, &pod)
	pod.Kind = "Pod"
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{Name:"pause", Image:"google/pause"})

	cpuRequest := un.Object["spec"].(map[string]interface{})["requests"].(map[string]interface{})["cpu"]
	if cpuRequest == nil {
		cpuRequest = "0m"
	}
	cpuQuantity, err := resource.ParseQuantity(cpuRequest.(string))
	if err != nil {
		klog.Errorf("parse cpu quantity error: %s", err)
		panic(err)
	}

	memRequest := un.Object["spec"].(map[string]interface{})["requests"].(map[string]interface{})["memory"]
	if memRequest == nil {
		memRequest = "0Mi"
	}
	memQuantity, err := resource.ParseQuantity(memRequest.(string))

	if err != nil {
		klog.Errorf("parse memory quantity error: %s", err)

	}
	pod.Spec.Containers[0].Resources.Requests = make(map[v1.ResourceName]resource.Quantity)
	pod.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	pod.Spec.Containers[0].Resources.Requests["memory"] = memQuantity

	pod.ResourceVersion = ""
	pod.Annotations["kind"] = GetPluralForm(un.GetKind())
	pod.Annotations["proxy"] = string(jsonBytes)
	pod.Spec.SchedulerName = "crdscheduler"
	return pod

}


func IsNullString(s string) bool{
	if s == "" {
		return true
	}
	return false
}


func GetPluralForm(s string) string {
	return strings.ToLower(s) + "s"
}
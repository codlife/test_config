package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"strconv"
	"syscall"

	//"k8s.io/kubernetes/pkg/apis/core"
	"os"
	"strings"
	"time"
)

// App contains deployed apps' info
type App struct {
	name      string
	pid       int
	shellFile string
}

type AppSpec struct {
	shellFile string
}


//typeNmme -> map[string][App]

var apps = make(map[string]map[string]App)

var file = "/etc/kubernetes/admin.conf"
var hostname = getHostName()
var dynamicClient dynamic.Interface
var client *kubernetes.Clientset

func main() {
	//apps := make(map[string]App)
	for {
		for _, resourceName := range GetCustomResourceNamesFromConfigMap() {
			go Loop(resourceName)
			time.Sleep(time.Second * 3)
		}
	}



}

func NewDynamicClient() (dynamic.Interface, error) {
	if dynamicClient != nil {
		return dynamicClient, nil
	}
	bytes, _ := ioutil.ReadFile(file)
	config, _ := clientcmd.NewClientConfigFromBytes(bytes)
	clientConfig, _ := config.ClientConfig()
	return dynamic.NewForConfig(clientConfig)
}

func NewClient() (*kubernetes.Clientset, error) {
	if client != nil {
		return client, nil
	}
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

func getHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}



func Loop(resourceName string) {


	dynamicClient, err := NewDynamicClient()

	if err != nil {
		panic(err)
	}
	client, err := NewClient()

	if err != nil {
		panic(err)
	}

	var gvr = schema.GroupVersionResource {
		Group:    "app.example.com",
		Version:  "v1alpha1",
		Resource: resourceName,
	}

	list, err := dynamicClient.Resource(gvr).
		Namespace("default").List(metav1.ListOptions{})
	jsonByte, err := json.Marshal(list)
	var data map[string]interface{}
	err = json.Unmarshal(jsonByte, &data)
	items := data["items"].([]interface{})
	if err != nil {
		panic(err)
	}
	appNames := make(map[string]AppSpec, 0)
	for i := range items {
		item := items[i].(map[string]interface{})
		fmt.Println(item)

		metadata := item["metadata"].(map[string]interface{})
		appName := metadata["name"].(string)
		fmt.Println("item: ", appName)

		spec := item["spec"].(map[string]interface{})
		var nodeName string

		switch spec["nodeName"].(type) {
		case nil:
		case string:
			nodeName = spec["nodeName"].(string)
		}
		switch spec["nodeName"].(type) {
		case nil:
		case string:
			nodeName = spec["nodeName"].(string)
		}
		fmt.Println("nodeName: ", nodeName)

		shellFile := spec["shellFile"].(string)
		fmt.Println("shellFile: ", shellFile)
		// if the nodeName in spec is equal to hostname, then add the appName into the appNames
		if strings.Compare(hostname, nodeName) == 0 {
			appNames[appName] = AppSpec{
				shellFile: shellFile,
			}
		}


	}

	fmt.Print("appNames")
	fmt.Println( len(appNames))
	fmt.Print("apps")
	fmt.Println(len(apps[resourceName]))
	for name := range apps[resourceName] {
		fmt.Printf("apps: %s\n", name)
		if _, ok := appNames[name]; !ok {
			pid := apps[resourceName][name].pid
			killProcessByPID(pid)
			delete(apps[resourceName], name)
			fmt.Println("Killed", pid)
			client.CoreV1().Pods(resourceName + "-ns").Delete(name, &metav1.DeleteOptions{})
		}
	}

	// check the app objects added by k8s, start the corresponding processes in the node
	// and add app's info in apps
	for name := range appNames {
		fmt.Printf("appNames: %s\n", name)
		if(apps[resourceName] == nil) {
			apps[resourceName] = make(map[string]App)
		}
		if _, ok := apps[resourceName][name]; !ok {
			// tomcatPath := "/usr/local/tomcat/apache-tomcat-8.5.35/bin/"
			// cmd := exec.Command("sh", tomcatPath+"startup.sh")
			cmd := exec.Command("sh", appNames[name].shellFile)
			err = cmd.Run() // it must be Run() rather than Start() here!!! Start() does not wait for it to complete!!!
			if err != nil {
				panic(err)
			}

			pid := getPIDByName("tomcat")
			fmt.Printf("Start tomcat: %d\n", pid)

			apps[resourceName][name] = App{
				name:      name,
				pid:       pid,
				shellFile: appNames[name].shellFile,
			}
		}
	}
}


func getPIDByName(name string) int {
	cmd := exec.Command("pgrep", "-f", name)
	// cmd = exec.Command("sh", "-c", `pgrep -f tomcat >> tmp`)
	// cmd = exec.Command("sh", "-c", `ps axf | grep tomcat | grep -v grep | awk '{print $1}'`)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	pid, err := strconv.Atoi(strings.Trim(string(out), "\n"))
	if err != nil {
		panic(err)
	}

	return pid
}

func killProcessByPID(pid int) {
	fmt.Printf("Kill process: %d\n", pid)
	err := syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		panic(err)
	}
}
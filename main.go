package main

import (
	"bufio"
	"fmt"
	"os"

	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

func main() {
	// 解析到config
	config, err := clientcmd.BuildConfigFromFlags("", "admin.conf")
	if err != nil {
		panic(err.Error())
	}

	// 创建连接
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var inputStr, PodName, PodNamespace string
	// 读取输入想要驱逐的Pod
	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Please input the Name and Namespace of the Pod you want to evict.")
		inputStr, err = inputReader.ReadString('\n')
		if len(inputStr) == 1 || err != nil {
			continue
		}
		fmt.Sscan(inputStr, &PodName, &PodNamespace)
		if len(PodName) == 0 || len(PodNamespace) == 0 {
			continue
		}
		// 根据输入的Pod名字和Namespace驱逐Pod
		EvictPod(clientset, PodName, PodNamespace)
	}
}

// EvictPod 用来驱逐一个Pod
func EvictPod(client kubernetes.Interface, name, namespace string) {
	fmt.Println("start evict pod " + name + "under" + namespace)

	eviction := &policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			Kind: "Eviction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	err := client.Policy().Evictions(eviction.Namespace).Evict(eviction)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	pod, _ := client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	eventRecorder := CreateEventRecorder(client)
	eventRecorder.Eventf(pod, v1.EventTypeWarning, "Evicted", "Eviction")

	fmt.Println("Evict pod " + pod.GetName() + "under" + namespace + " end.")
}

// CreateEventRecorder 创建一个用来发送event的接口对象
func CreateEventRecorder(client kubernetes.Interface) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "evict-go"})
}

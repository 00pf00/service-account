package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
)

func main() {
	logfile := flag.String("log", "/home/serviceaccount.log", "")
	flag.Parse()
	//初始化日志
	file, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE, 666);
	if err != nil{
		fmt.Println("open file fail")
	}
	//if err != nil {
	//	os.Exit(1)
	//}
	//logger := log.New(file,time.Now().String(),log.Ldate|log.Ltime|log.Lshortfile);

	//初始化client-go
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("rest.InClusterConfig()")
		file.WriteString("初始化client-go失败")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("kubernetes.NewForConfig(config)")
	}
	for {
		cm, err := clientset.CoreV1().ConfigMaps("tinykube").Get(context.TODO(), "codedns", metav1.GetOptions{})
		if err != nil {
			fmt.Println("get configmap fail")
			file.WriteString("get configmap fail ")
		}
		file.WriteString(cm.GetName())
		fmt.Println(cm.GetName())
		time.Sleep(1 * time.Second)
	}

}

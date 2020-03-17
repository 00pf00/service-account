package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
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
		cm, err := clientset.CoreV1().ConfigMaps("tinykube").Get(context.TODO(), "coredns", metav1.GetOptions{})
		if err != nil {
			fmt.Println("get configmap fail")
			fmt.Println(err)
			file.WriteString("get configmap fail ")
		}
		file.WriteString(cm.GetName())
		fmt.Println(cm.GetName())
		for k, v := range cm.Data {
			fmt.Printf("key = %s value = %s", k, v)
		}
		eps, err := clientset.CoreV1().Endpoints("tinykube").Get(context.TODO(), "proxycloud", metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("eps")
		for _, v := range eps.Subsets {
			for _, vv := range v.Addresses {
				fmt.Println(vv.IP)
			}
		}
		fmt.Println("ips")
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			fmt.Println(net.InterfaceAddrs())
		}
		for _, v := range addrs {
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					fmt.Println(ipnet.IP.To4())
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

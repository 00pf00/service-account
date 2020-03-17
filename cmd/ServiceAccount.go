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
	"strings"
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
		//获取configmap
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
		//更新configmap
		core := cm.Data["Corefile"]
		start := strings.Index(core, "}")
		var hosts string
		if strings.Contains(core, "hosts") {
			hosts += core[:start+2]
			d := core[start+2:]
			e := strings.Index(core[start+2:], "}")
			hosts += d[e+2:]
		} else {
			hosts += core[:start+2]
			hosts += "    hosts {\n"
			hosts += "        127.0.0.1     localhost\n"
			hosts += "        fallthrough\n"
			hosts += "    }\n"
			hosts += core[start+2:]
		}
		cm.Data["Corefile"] = hosts
		ucm, err := clientset.CoreV1().ConfigMaps("tinykube").Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("update configmap fail")
		}

		for k, v := range ucm.Data {
			fmt.Printf("updated key = %s value = %s", k, v)
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

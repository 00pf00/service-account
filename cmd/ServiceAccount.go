package main

import (
	"00pf00/service-account/pkg/util"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"strconv"
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
	//获取namespace
	nf, err := os.Open(util.NAMESPACEPATH)
	if err != nil {
		fmt.Println("open namespace file fail ")
	}
	nsb, err := ioutil.ReadAll(nf)
	if err != nil {
		fmt.Println("read namespce fail !")
	}
	nss := string(nsb);
	fmt.Println(nss)

	fmt.Println("ips")
	var h string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(net.InterfaceAddrs())
	}
	for _, v := range addrs {
		if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				h = ipnet.IP.To4().String()
			}
		}
	}
	var hs [10]string
	for i := 0; i < 10; i++ {
		hs[i] = h + "-" + strconv.Itoa(i)
	}
	order := 10
	count := 0
	for {

		 epss := []string{}
		//eps
		eps, err := clientset.CoreV1().Endpoints(nss).Get(context.TODO(), "proxycloud", metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("eps")
		for _, v := range eps.Subsets {
			for _, vv := range v.Addresses {
				fmt.Println(vv.IP)
				epss = append(epss,vv.IP )
			}
		}

		//获取clusterip
		s, err := clientset.CoreV1().Services(nss).Get(context.TODO(), util.SERVICENAME, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
			fmt.Println("get service fail")
		}
		sc := s.Spec.ClusterIP
		fmt.Printf(sc)
		//获取configmap
		cm, err := clientset.CoreV1().ConfigMaps(nss).Get(context.TODO(), "coredns", metav1.GetOptions{})
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
			d := core[start:+2]
			fmt.Println(string(d))
			s := strings.Index(d, "{")
			fmt.Println(s)
			e := strings.Index(d, "}")
			fmt.Println(e)
			hosts += d[:s]
			hss := strings.Split(string(d[s+1:e-1]), "\n")
			for _, v := range hss {
				if strings.Contains(v, h)  || strings.Contains(v,"fallthrough"){
					continue
				}
				flag := false
				for _,vv := range epss {
					if strings.Contains(v,vv){
						flag = true
						break
					}
				}
				if flag {
					hosts += v+"\n"
				}
			}
			for _, v := range hs {
				hosts += "      " + v + "     " + h + "\n"
			}
			hosts += "      fallthrough\n"
			hosts += d[e:]
		} else {
			hosts += core[:start+2]
			hosts += "    hosts {\n"
			hosts += "        127.0.0.1     localhost\n"
			hosts += "        fallthrough\n"
			hosts += "    }\n"
			hosts += core[start+2:]
		}
		cm.Data["Corefile"] = hosts
		ucm, err := clientset.CoreV1().ConfigMaps(nss).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println("update configmap fail")
		}

		for k, v := range ucm.Data {
			fmt.Printf("updated key = %s value = %s", k, v)
		}
		hs[count] = h + "-" + strconv.Itoa(order)
		order += 1
		if count == 9 {
			count = 0
		} else {
			count += 1
		}
		time.Sleep(10 * time.Second)
	}
}

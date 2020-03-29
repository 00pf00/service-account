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

const (
	COREFILE = "/etc/proxy/Corefile"
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
	var hs []string
	for i := 0; i < 10; i++ {
		hs = append(hs, h + "-" + strconv.Itoa(i))
	}
	order := 10
	count := 0
	for {

		//eps
		epss := []string{}
		eps, err := clientset.CoreV1().Endpoints(nss).Get(context.TODO(), util.SERVICENAME, metav1.GetOptions{})
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
		if CheckCM(h,epss,hs){
			updateCM(clientset, nss, h, epss, hs)
		}
		hs[count] = h + "-" + strconv.Itoa(order)
		order += 1
		if count == 9 {
			count = 0
		} else {
			count += 1
		}
		time.Sleep(20 * time.Second)
	}
}
func updateCM(clientset *kubernetes.Clientset, nss, h string, epss, hs []string) {
	//获取configmap
	cm, err := clientset.CoreV1().ConfigMaps(nss).Get(context.TODO(), "coredns", metav1.GetOptions{})
	if err != nil {
		fmt.Println("get configmap fail")
		fmt.Println(err)
	}
	fmt.Println(cm.GetName())
	for k, v := range cm.Data {
		fmt.Printf("key = %s value = %s", k, v)
	}
	//更新configmap
	core := cm.Data["Corefile"]
	var hosts string
	if strings.Contains(core, "hosts") {
		start := strings.Index(core, "hosts")
		hosts += core[:start]
		d := core[start:]
		s := strings.Index(d, "{")
		e := strings.Index(d, "}")
		hosts += d[:s+2]
		hss := strings.Split(string(d[s+2:e-1]), "\n")
		for _, v := range hss {
			if strings.Contains(v, h) || strings.Contains(v, "fallthrough") {
				continue
			}
			flag := false
			for _, vv := range epss {
				if strings.Contains(v, vv) {
					flag = true
					break
				}
			}
			if flag {
				hosts += v + "\n"
				vs := strings.Split(v, " ")
				for k, v := range vs {
					fmt.Printf("k = %v v = %v\n", k, v)
				}
			}
		}
		for _, v := range hs {
			hosts += "      " + h + "     " + v + "\n"
		}
		hosts += "      fallthrough\n    "
		hosts += d[e:]
	} else {
		start := strings.Index(core, "}")
		hosts += core[:start+2]
		hosts += "    hosts {\n"
		for _, v := range hs {
			hosts += "      " + h + "     " + v + "\n"
		}
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
}

func CheckCM(h string, epss,  hs []string) bool {
	cbs, err := ioutil.ReadFile(COREFILE)
	if err != nil {
		fmt.Printf("read corefile fail! err = %v", err)
		return false
	}
	core := string(cbs)
	if strings.Contains(core, "hosts") {
		start := strings.Index(core, "hosts")
		d := core[start:]
		s := strings.Index(d, "{")
		e := strings.Index(d, "}")
		hss := strings.Split(string(d[s+2:e-1]), "\n")

		contain := false
		for _, hssv := range hss {
			ks := []string{}
			if strings.Contains(hssv, h) {
				if !contain {
					contain = true
				}
				kss := strings.Split(hssv, " ")
				fmt.Printf("length = %d\n",len(kss))
				for _, kssv := range kss {
					fmt.Println(kssv)
					if kssv == "" {
						continue
					}
					fmt.Printf("append ks = %s\n",kssv)
					fmt.Printf("append ks length = %d\n",len(kssv))
					ks = append(ks, kssv)
				}
			}else {
				continue
			}
			flag := true
			for _, hsv := range hs {
				fmt.Printf("hsv = %s\n",hsv)
				fmt.Printf("ks[1] = %s\n",ks[1])
				if hsv == ks[1] {
					flag = false
					break
				}
			}
			if flag {
				fmt.Println("A not contain\n")
				return true
			}
		}
		if !contain {
			fmt.Println("B not contain\n")
			return true
		}
	}
	return false
}

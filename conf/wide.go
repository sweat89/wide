package conf

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type Conf struct {
	Server                string
	StaticServer          string
	EditorChannel         string
	OutputChannel         string
	ShellChannel          string
	StaticResourceVersion string
	ContextPath           string
	StaticPath            string
	RuntimeMode           string
	ProjectHome           string
}

var Wide Conf

func init() {
	flag.Set("logtostderr", "true")

	flag.Parse()

	bytes, _ := ioutil.ReadFile("conf/wide.json")

	err := json.Unmarshal(bytes, &Wide)
	if err != nil {
		glog.Errorln(err)
		return
	}

	ip := getNetworkInterface()
	glog.Infof("IP [%s]", ip)

	Wide.Server = strings.Replace(Wide.Server, "{IP}", ip, 1)
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{IP}", ip, 1)

	Wide.EditorChannel = strings.Replace(Wide.EditorChannel, "{IP}", ip, 1)
	Wide.OutputChannel = strings.Replace(Wide.OutputChannel, "{IP}", ip, 1)
	Wide.ShellChannel = strings.Replace(Wide.ShellChannel, "{IP}", ip, 1)

	glog.Info("Conf: \n" + string(bytes))
}

func getNetworkInterface() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}

	ret := "0.0.0.0"
	for _, addr := range addrs {
		if "0.0.0.0" != addr.String() {
			return addr.String()
		}
	}

	return ret
}

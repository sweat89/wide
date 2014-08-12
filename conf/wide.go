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
	ServerPort            string
	StaticServerScheme    string
	StaticServerHost      string
	StaticServerPort      string
	EditorChannel         string
	OutputChannel         string
	ShellChannel          string
	StaticResourceVersion string
	ContextPath           string
	StaticPath            string
	RuntimeMode           string
	ProjectHome           string

	StaticServer string
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

	Wide.StaticServer = ""
	if "" != Wide.StaticServerHost {
		Wide.StaticServer = Wide.StaticServerScheme +
			"://" + Wide.StaticServerHost +
			":" + Wide.StaticServerPort +
			Wide.StaticPath
	}

	ip := getNetworkInterface(1)
	glog.Infof("IP [%s]", ip)

	Wide.EditorChannel = strings.Replace(Wide.EditorChannel, "{IP}", ip, 1)
	Wide.OutputChannel = strings.Replace(Wide.OutputChannel, "{IP}", ip, 1)
	Wide.ShellChannel = strings.Replace(Wide.ShellChannel, "{IP}", ip, 1)

	glog.Info("Conf: \n" + string(bytes))
}

func getNetworkInterface(idx int) string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}

	return addrs[idx].String()
}

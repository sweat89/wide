package conf

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"io/ioutil"
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

	glog.Info("Conf: \n" + string(bytes))
}

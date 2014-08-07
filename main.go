package main

import (
	"encoding/json"
	"github.com/88250/wide/conf"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const PATH_SEPARATOR = string(os.PathSeparator)

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, map[string]string{"StaticServer": conf.Wide.StaticServer})
}

func run(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var c interface{}

	if err := decoder.Decode(&c); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	m := c.(map[string]interface{})

	projectName := m["project"].(string)
	projectPath := conf.Wide.ProjectHome + PATH_SEPARATOR + projectName
	filePath := projectPath + PATH_SEPARATOR + m["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := m["code"].(string)

	glog.Info(string(code[192]))

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	argv := []string{"run", filePath}

	cmd := exec.Command("go", argv...)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	reader := io.MultiReader(stdout, stderr)

	cmd.Start()

	rec := map[string]interface{}{}

	for {
		buf := make([]byte, 1024)
		count, err := reader.Read(buf)
		if nil != err || 0 == count {
			break
		} else {
			rec["output"] = string(buf)

			err := outputWS.WriteJSON(&rec)
			if nil != err {
				glog.Error(err)
				break
			}
		}
	}

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func fmt(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var c interface{}

	if err := decoder.Decode(&c); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	m := c.(map[string]interface{})

	projectName := m["project"].(string)
	projectPath := conf.Wide.ProjectHome + PATH_SEPARATOR + projectName
	filePath := projectPath + PATH_SEPARATOR + m["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := m["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	argv := []string{filePath}

	cmd := exec.Command("gofmt", argv...)

	bytes, _ := cmd.Output()
	output := string(bytes)

	succ := true
	if "" == output {
		succ = false
	}

	ret, _ := json.Marshal(map[string]interface{}{"succ": succ, "code": string(output)})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

// TODO: 多会话支持
var outputWS *websocket.Conn

func output(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)

	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		glog.Error(err)
		return
	}

	outputWS = ws

	ret := map[string]interface{}{"output": "Ouput 初始化完毕\n"}
	outputWS.WriteJSON(&ret)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", index)
	http.HandleFunc("/run", run)
	http.HandleFunc("/fmt", fmt)

	http.HandleFunc("/output", output)

	err := http.ListenAndServe(":"+conf.Wide.ServerPort, nil)
	if err != nil {
		glog.Fatal(err)
	}
}

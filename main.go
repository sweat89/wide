package main

// TODO:
// 1. 项目支持
// 2. 编辑锁
// 3. Shell
//

import (
	"bytes"
	"encoding/json"
	"github.com/88250/wide/conf"
	"github.com/88250/wide/files"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const PATH_SEPARATOR = string(os.PathSeparator)

var sessionStore = sessions.NewCookieStore([]byte("BEYOND"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "wide-session")

	if session.IsNew {
		session.Values["id"] = strconv.Itoa(rand.Int())
	}

	session.Save(r, w)

	t, err := template.ParseFiles("templates/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, map[string]interface{}{"Wide": conf.Wide})
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	projectName := args["project"].(string)
	projectPath := conf.Wide.ProjectHome + PATH_SEPARATOR + projectName
	filePath := projectPath + PATH_SEPARATOR + args["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	projectName := args["project"].(string)
	projectPath := conf.Wide.ProjectHome + PATH_SEPARATOR + projectName
	filePath := projectPath + PATH_SEPARATOR + args["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)

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

	go func() {
		session, _ := sessionStore.Get(r, "wide-session")
		sid := session.Values["id"].(string)

		for {
			buf := make([]byte, 1024)
			count, err := reader.Read(buf)

			if nil != err || 0 == count {
				break
			} else {
				rec["output"] = string(buf[:count])

				if nil != outputWS[sid] {
					err := outputWS[sid].WriteJSON(&rec)
					if nil != err {
						glog.Error(err)
						break
					}
				}
			}
		}
	}()

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func fmtHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	projectName := args["project"].(string)
	projectPath := conf.Wide.ProjectHome + PATH_SEPARATOR + projectName
	filePath := projectPath + PATH_SEPARATOR + args["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)

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

var outputWS = map[string]*websocket.Conn{}
var editorWS = map[string]*websocket.Conn{}
var shellWS = map[string]*websocket.Conn{}

func outputHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	outputWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Ouput initialized\n", "cmd": "init-output"}
	outputWS[sid].WriteJSON(&ret)

	glog.Info("Output channels: ", len(outputWS))
}

func editorWSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	editorWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Editor initialized", "cmd": "init-editor"}
	editorWS[sid].WriteJSON(&ret)

	glog.Info("Editor channels: ", len(outputWS))

	args := map[string]interface{}{}
	for {
		if err := editorWS[sid].ReadJSON(&args); err != nil {
			if err.Error() == "EOF" {
				return
			}
			// ErrShortWrite means that a write accepted fewer bytes than requested but failed to return an explicit error.
			if err.Error() == "unexpected EOF" {
				return
			}

			glog.Error("Editor WS ERROR: " + err.Error())
			return
		}

		code := args["code"].(string)
		line := int(args["cursorLine"].(float64))
		ch := int(args["cursorCh"].(float64))

		offset := getCursorOffset(code, line, ch)

		// glog.Infof("offset: %d", offset)

		argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}

		var output bytes.Buffer

		cmd := exec.Command("gocode", argv...)
		cmd.Stdout = &output

		stdin, _ := cmd.StdinPipe()
		cmd.Start()
		stdin.Write([]byte(code))
		stdin.Close()
		cmd.Wait()

		ret = map[string]interface{}{"output": string(output.Bytes()), "cmd": "autocomplete"}

		if err := editorWS[sid].WriteJSON(&ret); err != nil {
			glog.Error("Editor WS ERROR: " + err.Error())
			return
		}
	}
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)
	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	// glog.Infof("offset: %d", offset)

	argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}

	var output bytes.Buffer

	cmd := exec.Command("gocode", argv...)
	cmd.Stdout = &output

	stdin, _ := cmd.StdinPipe()
	cmd.Start()
	stdin.Write([]byte(code))
	stdin.Close()
	cmd.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.Write(output.Bytes())

}

func shellWSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	shellWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Shell initialized", "cmd": "init-shell"}
	shellWS[sid].WriteJSON(&ret)

	glog.Info("Shell channels: ", len(outputWS))

	input := map[string]interface{}{}

	for {
		if err := shellWS[sid].ReadJSON(&input); err != nil {
			if err.Error() == "EOF" {
				return
			}
			// ErrShortWrite means that a write accepted fewer bytes than requested but failed to return an explicit error.
			if err.Error() == "unexpected EOF" {
				return
			}

			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}

		var output bytes.Buffer

		inputCmd := input["cmd"].(string)

		cmdWithArgs := strings.Split(inputCmd, " ")

		cmd := exec.Command(cmdWithArgs[0], cmdWithArgs[:1]...)
		cmd.Stdout = &output

		cmd.Start()
		cmd.Wait()

		ret = map[string]interface{}{"output": string(output.Bytes()), "cmd": "shell-output"}

		if err := shellWS[sid].WriteJSON(&ret); err != nil {
			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}
	}
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}

	root := files.FileNode{"projects", conf.Wide.ProjectHome, "d", []*files.FileNode{}}
	fileInfo, _ := os.Lstat(conf.Wide.ProjectHome)

	files.Walk(conf.Wide.ProjectHome, fileInfo, &root)

	data["root"] = root

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func getFile(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	path := args["path"].(string)
	extension := path[strings.LastIndex(path, "."):]

	buf, _ := ioutil.ReadFile(path)

	content := string(buf)

	data := map[string]interface{}{"succ": true}
	data["content"] = content
	data["mode"] = getEditorMode(extension)

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func getEditorMode(filenameExtension string) string {
	switch filenameExtension {
	case ".go":
		return "go"
	case ".html":
		return "htmlmixed"
	case ".md":
		return "markdown"
	case ".js", ".json":
		return "javascript"
	case ".css":
		return "css"
	case ".xml":
		return "xml"
	default:
		return "text"
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/fmt", fmtHandler)

	http.HandleFunc("/files", getFiles)
	http.HandleFunc("/file", getFile)

	http.HandleFunc("/output/ws", outputHandler)
	http.HandleFunc("/editor/ws", editorWSHandler)
	http.HandleFunc("/shell/ws", shellWSHandler)

	http.HandleFunc("/autocomplete", autocompleteHandler)

	err := http.ListenAndServe(":"+conf.Wide.ServerPort, nil)
	if err != nil {
		glog.Fatal(err)
	}
}

// 工具

func getCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	offset += line + ch

	return
}

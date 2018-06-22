package sender

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Receiver struct {
	Name string   `json:"name"`
	Port int      `json:"port"`
	Ext  []string `json:"ext"`
}

type SendFileInfo struct {
	Filename   string    `json:"filename"`
	Content    string    `json:"content"`
	ModifiedAt time.Time `json:"modifiedat"`
}

const (
	folderToCheck  = "ToCheck\\"
	checkfileroute = "checkfile"
	filetransfer   = "filetransfer"
	NotModified    = 888
	Modified       = 999
)

var mapReceiver = map[string][]int{}

// ************** HANDLERS ************** //

// HandlerRegister register receivers ports and file extension
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	var rec Receiver
	decodeJSON(r, &rec)
	for _, ext := range rec.Ext {
		ports, ok := mapReceiver[ext]
		if !ok {
			ports = make([]int, 0)
		}
		ports = append(ports, rec.Port)
		mapReceiver[ext] = ports
	}
	checkFiles(rec.Port)
}

func checkFiles(port int) {
	files := getDirContent()
	for _, file := range files {
		jsonReader := fileToReader(file, false)
		resp, err := http.Post("http://localhost:"+strconv.Itoa(port)+"/"+checkfileroute, "json", jsonReader)
		if err != nil {
			fmt.Printf("Could not send file info to server: %v", err)
		}

		if resp.StatusCode == Modified {
			sendFile(file)
		}
		fmt.Printf("file: %v, resp: %v\n", file.Name(), resp.StatusCode)
	}
}

// ************** PRIVATE FUNCIONS ************** //

func getDirContent() []os.FileInfo {
	files, err := ioutil.ReadDir(folderToCheck)
	if err != nil {
		fmt.Printf("Could not read files from directory: %v", err)
	}
	return files
}

func fileToReader(filename os.FileInfo, addContent bool) io.Reader {
	var strToBase64 string
	if addContent == true {
		content, err := ioutil.ReadFile(folderToCheck + filename.Name())
		if err != nil {
			fmt.Printf("Could not read from file: %v", err)
		}
		strToBase64 = base64.StdEncoding.EncodeToString(content)
	}
	jsonStruct := SendFileInfo{Filename: filename.Name(), Content: strToBase64, ModifiedAt: filename.ModTime()}
	jsonByte, err := json.Marshal(jsonStruct)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}
	jsonReader := bytes.NewReader(jsonByte)

	return jsonReader
}

func sendFile(filename os.FileInfo) *http.Response {
	jsonReader := fileToReader(filename, true)
	resp, err := http.Post("http://localhost:8888/"+filetransfer, "json", jsonReader)
	if err != nil {
		fmt.Printf("Could not send the file: %v", err)
	}

	return resp
}

func decodeJSON(r *http.Request, container interface{}) {
	rawJSON := json.NewDecoder(r.Body)
	err := rawJSON.Decode(container)
	if err != nil {
		fmt.Printf("Failed to decode the file: %v", err)
	}
	defer r.Body.Close()
	fmt.Println(container)
}

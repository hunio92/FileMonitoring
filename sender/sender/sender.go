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
	"strings"
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

type mapFile map[string]SendFileInfo

var fileMap mapFile

const (
	folderToCheck  = "ToCheck/"
	checkfileroute = "checkfile"
	filetransfer   = "filetransfer"
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
	ReceiverPort := strconv.Itoa(rec.Port)
	checkFiles(ReceiverPort)
}

// StoreFilesInfo store the current files info (name, last time modified)
func StoreFilesInfo() {
	fileMap = make(mapFile)
	files := getDirContent()
	for _, filename := range files {
		fileMap[filename.Name()] = SendFileInfo{Filename: filename.Name(), ModifiedAt: filename.ModTime()}
	}
}

func checkFiles(ReceiverPort string) {
	files := getDirContent()
	for _, file := range files {
		filename := file.Name()
		index := strings.Index(filename, ".")
		ports := mapReceiver[filename[index+1:]]
		for _, port := range ports {
			portStr := strconv.Itoa(port)
			if portStr == ReceiverPort {
				jsonData := fileToReader(file, false)
				resp := sendFile(portStr, checkfileroute, jsonData)

				if resp.StatusCode == http.StatusNotFound {
					jsonData := fileToReader(file, true)
					sendFile(ReceiverPort, filetransfer, jsonData)
				}
				fmt.Printf("file: %v, resp: %v\n", file.Name(), resp.StatusCode)
			}
		}
	}

	for {
		files := getDirContent()
		for _, file := range files {
			checkModified(file)
		}
		time.Sleep(3 * time.Second)
	}
}

// ************** PRIVATE FUNCIONS ************** //

func checkModified(file os.FileInfo) {
	filename := file.Name()
	index := strings.Index(filename, ".")
	ports := mapReceiver[filename[index+1:]]
	for _, port := range ports {
		portStr := strconv.Itoa(port)
		if fileMap[file.Name()].ModifiedAt != file.ModTime() {
			fmt.Printf("file: %v, port: %v", file.Name(), portStr)
			jsonReader := fileToReader(file, true)
			sendFile(portStr, filetransfer, jsonReader)
		}
	}
	fileMap[file.Name()] = SendFileInfo{ModifiedAt: file.ModTime()}
}

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

func sendFile(port string, route string, jsonData io.Reader) *http.Response {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Post("http://127.0.0.1:"+port+"/"+route, "json", jsonData)
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
}

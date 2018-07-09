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

	"github.com/spf13/viper"
)

type Receiver struct {
	Name      string   `json:"name"`
	Port      int      `json:"port"`
	Ext       []string `json:"ext"`
	Senderkey string   `json:"senderkey"`
}

type SendFileInfo struct {
	Filename   string    `json:"filename"`
	Content    string    `json:"content"`
	ModifiedAt time.Time `json:"modifiedat"`
	SenderKey  string    `json:"senderkey"`
}

const (
	folderToCheck     = "ToCheck/"
	checkFileRoute    = "checkfile"
	fileTransferRoute = "filetransfer"
)

var fileMap = map[string]SendFileInfo{}
var fileAuthKey = map[int]string{}
var mapReceiver = map[string][]int{}
var errorCounter = map[string]int{}

// ************** HANDLERS ************** //

// HandlerRegister register receivers ports and file extension
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("authkey")

	hasAccess := authentification(key)

	if hasAccess == true {
		var rec Receiver
		decodeJSON(r, &rec)
		fileAuthKey[rec.Port] = rec.Senderkey
		fmt.Println("senderkey: ", rec.Senderkey)
		for _, ext := range rec.Ext {
			ports, ok := mapReceiver[ext]
			if !ok {
				ports = make([]int, 0)
			}
			ports = append(ports, rec.Port)
			mapReceiver[ext] = ports
		}
		ReceiverPort := strconv.Itoa(rec.Port)
		fmt.Println(mapReceiver)
		checkFiles(ReceiverPort)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// ************** PUBLIC FUNCIONS ************** //

func CheckConnection() {
	var portToDelete int
	for {
		portToDelete = 0
		for k, v := range errorCounter {
			if v > 3 {
				fmt.Printf("The port: %v, failed: %v times \n", k, v)
				val, err := strconv.Atoi(k)
				if err != nil {
					fmt.Printf("Could not convert string to int: %v\n", err)
				}
				portToDelete = val
			}
		}
		for k, v := range mapReceiver {
			a := v
			i := 0
			for i < len(a) {
				if a[i] == portToDelete {
					a = append(a[:i], a[i+1:]...)
					mapReceiver[k] = a
				}
				i++
			}
		}

		delete(errorCounter, strconv.Itoa(portToDelete))
		time.Sleep(5 * time.Second)
	}
}

// StoreFilesInfo store the current files info (name, last time modified)
func StoreFilesInfo() {
	files := getDirContent()
	for _, filename := range files {
		fileMap[filename.Name()] = SendFileInfo{Filename: filename.Name(), ModifiedAt: filename.ModTime()}
	}
}

// ************** PRIVATE FUNCIONS ************** //

func checkFiles(ReceiverPort string) {
	files := getDirContent()
	for _, file := range files {
		filename := file.Name()
		index := strings.Index(filename, ".")
		ports := mapReceiver[filename[index+1:]]
		for _, port := range ports {
			portStr := strconv.Itoa(port)
			if portStr == ReceiverPort {
				resp := isModified(portStr, file)
				if !resp {
					errorCounter[portStr]++
					fmt.Println("error counter: ", errorCounter)
				}

				fmt.Printf("file: %v\n", file.Name())
			}
		}
	}

	go func() {
		for {
			files := getDirContent()
			for _, file := range files {
				checkModified(file)
			}
			time.Sleep(3 * time.Second)
		}
	}()
}

func checkModified(file os.FileInfo) {
	fileName := file.Name()
	index := strings.Index(fileName, ".")
	ports := mapReceiver[fileName[index+1:]]
	for _, port := range ports {
		portStr := strconv.Itoa(port)
		if fileMap[fileName].ModifiedAt != file.ModTime() {
			fmt.Printf("file: %v, port: %v, pid: %v\n", fileName, portStr, os.Getegid())
			resp, isSent := sendFile(portStr, fileTransferRoute, file, true)
			if !isSent {
				errorCounter[portStr]++
				fmt.Println("error counter: ", errorCounter)
			}
			if resp.StatusCode == http.StatusNotFound {
				fmt.Println("Mismatch in auth key !")
			}
		}
	}
	fileMap[fileName] = SendFileInfo{ModifiedAt: file.ModTime()}
}

func getDirContent() []os.FileInfo {
	files, err := ioutil.ReadDir(folderToCheck)
	if err != nil {
		fmt.Printf("Could not read files from directory: %v", err)
	}
	return files
}

func fileToReader(fileName os.FileInfo, senderKey string, addContent bool) io.Reader {
	var strToBase64 string
	if addContent == true {
		content, err := ioutil.ReadFile(folderToCheck + fileName.Name())
		if err != nil {
			fmt.Printf("Could not read from file: %v", err)
		}
		strToBase64 = base64.StdEncoding.EncodeToString(content)
	}
	jsonStruct := SendFileInfo{Filename: fileName.Name(), Content: strToBase64, ModifiedAt: fileName.ModTime(), SenderKey: senderKey}
	jsonByte, err := json.Marshal(jsonStruct)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}
	jsonReader := bytes.NewReader(jsonByte)

	return jsonReader
}

func isModified(port string, file os.FileInfo) bool {
	resp, _ := sendFile(port, checkFileRoute, file, false)

	if resp.StatusCode != http.StatusOK {
		_, isSent := sendFile(port, fileTransferRoute, file, true)
		if !isSent {
			return false
		}
		if resp.StatusCode == http.StatusNotFound {
			fmt.Println("Mismatch in auth key !")
		}
	}
	return true
}

func sendFile(port string, route string, file os.FileInfo, addContent bool) (*http.Response, bool) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	intPort, _ := strconv.Atoi(port)
	jsonData := fileToReader(file, fileAuthKey[intPort], addContent)
	resp, err := client.Post("http://127.0.0.1:"+port+"/"+fileTransferRoute, "json", jsonData)
	if err != nil {
		return resp, false
	}
	return resp, true
}

func decodeJSON(r *http.Request, container interface{}) {
	rawJSON := json.NewDecoder(r.Body)
	err := rawJSON.Decode(container)
	if err != nil {
		fmt.Printf("Failed to decode the file: %v", err)
	}
	defer r.Body.Close()
}

func authentification(key string) bool {
	viper.SetConfigName("authkeys")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Could not read config file: %v", err)
	}
	authKeys := viper.GetStringSlice("keys")

	isFound := false
	for _, authkey := range authKeys {
		if authkey == key {
			isFound = true
			break
		}
	}

	return isFound
}

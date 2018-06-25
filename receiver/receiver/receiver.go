package receiver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

// ************** DECLARATION / INITIALIZATION / CONSTS ************** //

const (
	folderToSave = "SaveTo/"
)

type SendStructure struct {
	Filename   string    `json:"filename"`
	Content    string    `json:"content"`
	ModifiedAt time.Time `json:"modifiedat"`
}

type ReceiverConfig struct {
	Name string   `json:"name"`
	Port int      `json:"port"`
	Ext  []string `json:"ext"`
}

var Config ReceiverConfig

// ************** HANDLERS ************** //

// HandlerCheckFile if the file has been modified
func HandlerCheckFile(w http.ResponseWriter, r *http.Request) {
	var fileInfo SendStructure
	decodeJSON(r, &fileInfo)
	info := getFilesInfo()
	for _, file := range info {
		if fileInfo.Filename == file.Filename && fileInfo.ModifiedAt == file.ModifiedAt {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// HandlerSaveFile get file and save
func HandlerFileTransfer(w http.ResponseWriter, r *http.Request) {
	var fileInfo SendStructure
	decodeJSON(r, &fileInfo)
	msg, err := decodeMessage(fileInfo.Content)
	if err != nil {
		fmt.Printf("Could not decode the base64 message: %v", err)
	}

	writeFile(folderToSave+fileInfo.Filename, msg)
}

// ************** PUBLIC FUNCIONS ************** //

// SaveFilesInfo create json file and saves the files info (name, last modified)
func SaveFilesInfo() {
	jsonStruct := make([]SendStructure, 0)
	files := getDirContent(folderToSave)
	for _, filename := range files {
		jsonStruct = append(jsonStruct, SendStructure{Filename: filename.Name(), ModifiedAt: filename.ModTime()})
	}
	jsonByte, err := json.Marshal(jsonStruct)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}
	writeFile("filesinfo.json", jsonByte)
}

// ReadConfig read the receiver configurations (port, name, extensions)
func ReadConfig(configFile string) {
	viper.SetConfigName(configFile)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Could not read config file: %v", err)
	}
	err1 := viper.Unmarshal(&Config)
	if err1 != nil {
		fmt.Printf("Could not unmarshal data: %v", err)
	}
}

// ************** PRIVATE FUNCIONS ************** //

func PostData() {
	jsonByte, err := json.Marshal(Config)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}
	jsonReader := bytes.NewReader(jsonByte)
	_, err = http.Post("http://127.0.0.1:8080/register", "json", jsonReader)
	if err != nil {
		fmt.Printf("Could not send the file: %v", err)
	}
}

func decodeMessage(msg string) ([]byte, error) {
	buff, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, fmt.Errorf("Error to decode message: %d", buff)
	}
	return buff, nil
}

func decodeJSON(r *http.Request, container interface{}) {
	rawJSON := json.NewDecoder(r.Body)
	err := rawJSON.Decode(container)
	if err != nil {
		fmt.Printf("Failed to decode the file: %v", err)
	}
	defer r.Body.Close()
}

func getFilesInfo() (filesInfo []SendStructure) {
	raw, err := ioutil.ReadFile("filesinfo.json")
	if err != nil {
		fmt.Printf("Could not read json file: %q \n", raw)
	}
	err = json.Unmarshal(raw, &filesInfo)
	if err != nil {
		fmt.Printf("Could not convert json file: %q \n", raw)
	}
	return filesInfo
}

func getDirContent(path string) []os.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("Could not read files from directory: %v", err)
	}
	return files
}

func writeFile(filename string, msg []byte) {
	err := ioutil.WriteFile(filename, msg, 0200)
	if err != nil {
		fmt.Printf("Could not write to file: %v", err)
	}
}

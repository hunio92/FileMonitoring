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

// TestETCDOutage
// ************** DECLARATION / INITIALIZATION / CONSTS ************** //

type SendStructure struct {
	Filename   string    `json:"filename"`
	Content    string    `json:"content"`
	ModifiedAt time.Time `json:"modifiedat"`
	SenderKey  string    `json:"senderkey"`
}

type ReceiverConfig struct {
	Name      string   `json:"name"`
	Port      int      `json:"port"`
	Ext       []string `json:"ext"`
	Key       string   `json:"key"`
	Senderkey string   `json:"senderkey"`
}

var config ReceiverConfig
var receiverFolder = "DefaultPath/"

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

// HandlerFileTransfer get file from sender and save
func HandlerFileTransfer(w http.ResponseWriter, r *http.Request) {
	var fileInfo SendStructure
	decodeJSON(r, &fileInfo)
	fmt.Printf("senderkey: %v , receveerkey: %v \n", fileInfo.SenderKey, config.Senderkey)
	if fileInfo.SenderKey == config.Senderkey {
		msg, err := decodeMessage(fileInfo.Content)
		if err != nil {
			fmt.Printf("Could not decode the base64 message: %v", err)
		}
		writeFile(receiverFolder+fileInfo.Filename, msg)
		err = os.Chtimes(receiverFolder+fileInfo.Filename, time.Now(), fileInfo.ModifiedAt)
		if err != nil {
			fmt.Printf("Could not change the last time modified field: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// ************** PUBLIC FUNCIONS ************** //

func PostData() {
	jsonByte, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	jsonReader := bytes.NewReader(jsonByte)
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/register", jsonReader)
	if err != nil {
		fmt.Printf("Could not create the request: %v", err)
	}

	req.Header.Set("authkey", config.Key)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Could not send auth data: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Could not auth: %v", resp)
	}
}

// SaveFilesInfo create json file and saves the files info (name, last modified)
func SaveFilesInfo(receiverName string) {
	jsonStruct := make([]SendStructure, 0)
	files := getDirContent(receiverName + "/")
	for _, filename := range files {
		jsonStruct = append(jsonStruct, SendStructure{Filename: filename.Name(), ModifiedAt: filename.ModTime()})
	}
	jsonByte, err := json.Marshal(jsonStruct)
	if err != nil {
		fmt.Printf("Could not create JSON: %v", err)
	}
	writeFile("FilesInfo/"+receiverName+".json", jsonByte)
}

// ReadConfig read the receiver configurations (port, name, extensions)
func ReadConfig(configFile string) string {
	viper.SetConfigName(configFile)
	viper.AddConfigPath("./ReceiverConfig/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Could not read config file: %v", err)
	}
	err1 := viper.Unmarshal(&config)
	if err1 != nil {
		fmt.Printf("Could not unmarshal data: %v", err1)
	}

	fmt.Println(config)
	receiverFolder = config.Name + "/"
	_, err = os.Stat(receiverFolder)
	if err != nil {
		os.Mkdir(receiverFolder, 0600)
	}

	return config.Name
}

func GetConfig() ReceiverConfig {
	return config
}

// ************** PRIVATE FUNCIONS ************** //

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
	raw, err := ioutil.ReadFile("FilesInfo/" + config.Name + ".json")
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

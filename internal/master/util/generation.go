package util

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func ConvertFileContentsToBase64() {
	conversionFolderName := "conversions"
	files, _ := ioutil.ReadDir(conversionFolderName)

	for i := 0; i < len(files); i++ {
		if strings.Contains(files[i].Name(), "result") == false {
			fileData, _ := ioutil.ReadFile(path.Join(conversionFolderName, files[i].Name()))
			encodedData := base64.StdEncoding.EncodeToString(fileData)

			err := ioutil.WriteFile(path.Join(conversionFolderName, files[i].Name()+".result"), []byte(encodedData), os.ModeExclusive)

			if err != nil {
				fmt.Printf("Error converting contents of file %v. Error: %v", files[i].Name(), err)
			}
		}
	}
}

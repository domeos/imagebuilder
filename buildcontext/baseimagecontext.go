package buildcontext

import (
	"imagebuilder/buildfile"
	"fmt"
	"encoding/json"
	"os"
	"net/http"
	"io/ioutil"
	"strings"
)

type BaseImageContext  struct {
	Server      string     `json:"server"`

	ImageId     string     `json:"imageId"`

	ImageName   string     `json:"imageName"`

	ImageTag    string     `json:"imageTag"`

	RegistryUrl string     `json:"registryUrl"`

	Dockerfile  string     `json:"dockerfile"`

	Secret      string     `json:"secret"`

	CodeType    string     `json:"codeType"`
}

type FileInfo struct {
	FileName string
	FilePath string
	Md5      string
}

type FileSlice struct {
	Files []FileInfo
}

func writeFile(filecontent, filename string) error {
	fout, err := os.Create(LocalCodePath + filename)
	if err != nil {
		fmt.Println("create script file error,", err.Error())
		return err
	}
	defer fout.Close()
	fout.WriteString(filecontent)
	return nil
}

func (context *BaseImageContext) WriteBaseImageScript() (script string, error error) {
	f := buildfile.New()
	f.WriteCmdSilent(fmt.Sprintf("if [ ! -d \"%s\" ]; then mkdir %s; fi", LocalCodePath, LocalCodePath))
	f.WriteCmdSilent(fmt.Sprintf("cd %s", LocalCodePath))
	f.WriteCmdSilent(fmt.Sprintf("curl --connect-timeout 60 -o %s %s/api/image/custom/download/dockerfile/%s?secret=%s", LocalCodePath + "/Dockerfile", context.Server, context.ImageId, context.Secret))
	url := fmt.Sprintf("%s/api/image/custom/getfilejson/%s?secret=%s", context.Server, context.ImageId, context.Secret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	filejson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var fileinfos FileSlice
	json.Unmarshal(filejson, &fileinfos)
	for _, fileinfo := range fileinfos.Files {
		f.WriteCmdSilent(fmt.Sprintf("curl -o %s %s/api/image/custom/download/%s/%s?secret=%s", fileinfo.FileName, context.Server, context.ImageId, fileinfo.Md5, context.Secret))
	}

	imageInfo := ""

	if len(context.RegistryUrl) > 0 {
		imageInfo = strings.TrimPrefix(context.RegistryUrl, "http://")
		imageInfo = strings.TrimPrefix(imageInfo, "https://") + "/"
	}
	imageInfo = imageInfo + context.ImageName
	if len(context.ImageTag) > 0 {
		imageInfo = imageInfo + ":" + context.ImageTag
	} else {
		imageInfo = imageInfo + ":latest"
	}
	f.WriteCmd(fmt.Sprintf("docker build --pull -t %s .", imageInfo))
	f.WriteCmd(fmt.Sprintf("docker push %s", imageInfo))
	return f.String(), nil
}

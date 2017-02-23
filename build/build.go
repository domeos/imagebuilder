package build

import (
	"imagebuilder/buildcontext"
	"imagebuilder/contect"
	"net/http"
	"fmt"
	"encoding/json"
	"os"
	"bytes"
	"io"
	"mime/multipart"
	"io/ioutil"
	"time"
	"strings"
	"os/exec"
	"strconv"
)

var ScriptFile string = "/root/exec.sh"

const (
	Fail string = "Fail"
	Success string = "Success"
)

func writeScriptFile(script string) error {
	fout, err := os.Create(ScriptFile)
	if err != nil {
		fmt.Println("create script file error,", err.Error())
		return err
	}
	defer fout.Close()
	fout.WriteString(script)
	return nil
}

type buildStatus struct {
	ProjectId int `json:"projectId"`
	BuildId   int `json:"buildId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (bs *buildStatus)setBuildStatus(server string, secret string, projectType string, timeout time.Duration) error {
	text, err := json.Marshal(bs)
	if err != nil {
		fmt.Println("marshal build status json error,", err)
		return err
	} else {
		body := bytes.NewBuffer([]byte(text))
		client := http.Client{
			Timeout: timeout,
		}
		if (strings.EqualFold(projectType, "BASEIMAGECUSTOM")) {
			_, err = client.Post(server + "/api/image/custom/status?secret=" + secret, "application/json;charset=utf-8", body)
		} else {
			_, err = client.Post(server + "/api/ci/build/status?secret=" + secret, "application/json;charset=utf-8", body)
		}

		if err != nil {
			fmt.Println("send build status to server error,", err.Error())
			return err
		}
		return nil
	}
}

func (bs *buildStatus) UploadFile(url string, timeout time.Duration) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	//iocopy
	file, err := os.Open(contect.LogFilename);
	if err != nil {
		return err
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Post(url, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(resp_body))
	return nil
}

func execScript(script string, bs *buildStatus, bc string) {
	cmd := exec.Command("chmod", "755", ScriptFile)
	err := cmd.Run()
	if err != nil {
		bs.Status = Fail
		bs.Message = "generate ci script fail, parameter: " + bc
	} else {
		executor := contect.Executor{}
		err = executor.Init()
		if err != nil {
			bs.Status = Fail
			bs.Message = "init log file error"
		} else {
			err := executor.Command("/bin/sh", ScriptFile)
			//err = err.(syscall.Errno)
			if err != nil {
				bs.Status = Fail
				bs.Message = "execute ci script fail"
			} else {
				bs.Status = Success
			}
		}
		executor.Close()
	}
}

func uploadLogFile(bs *buildStatus, uploadUrl, server, secret, codeType string) {
	var uploadretry time.Duration = 0;
	var err error
	for {
		timeout := time.Duration(time.Second * (10 + 20 * uploadretry))
		err = bs.UploadFile(uploadUrl, timeout)
		if (err != nil) {
			bs.Message = "upload log file error"
			fmt.Println("upload log file error,", err.Error())
			uploadretry += 1
			if uploadretry >= 3 {
				break
			}
		} else {
			break
		}
	}
	uploadretry = 0;
	for {
		timeout := time.Duration(time.Second * (10 + 20 * uploadretry))
		err = bs.setBuildStatus(server, secret, codeType, timeout)
		if (err != nil) {
			fmt.Println("set build status error,", err.Error())
			uploadretry += 1
			if uploadretry >= 3 {
				break
			}
		} else {
			break
		}
	}
}

func RunOnType(codeType string) {
	switch codeType {
	case "BASEIMAGECUSTOM":
		server := os.Getenv("SERVER")
		imageId := os.Getenv("IMAGEID")
		imageName := os.Getenv("IMAGENAME")
		imageTag := os.Getenv("IMAGETAG")
		registryUrl := os.Getenv("REGISTRYURL")
		dockerfile := os.Getenv("DOCKERFILE")
		secret := os.Getenv("SECRET")
		buildContext := &buildcontext.BaseImageContext{server, imageId, imageName, imageTag, registryUrl, dockerfile, secret, codeType}
		bs := &buildStatus{}
		bc, _ := json.Marshal(buildContext)
		bs.ProjectId, _ = strconv.Atoi(imageId)
		script, err := buildContext.WriteBaseImageScript()
		if err != nil {
			bs.Status = Fail
			bs.Message = "generate ci script fail, parameter: " + string(bc)

		} else {
			err = writeScriptFile(script)
			if err != nil {
				bs.Status = Fail
				bs.Message = "write ci script fail, parameter: " + string(bc)
			} else {
				execScript(script, bs, string(bc))
			}
		}
		uploadUrl := fmt.Sprintf(server + "/api/image/custom/upload/%s?secret=%s", imageId, secret)
		uploadLogFile(bs, uploadUrl, server, secret, codeType)

	default:
		server := os.Getenv("SERVER")
		buildId, _ := strconv.Atoi(os.Getenv("BUILD_ID"))
		idrsa := os.Getenv("IDRSA")
		codeUrl := os.Getenv("CODE_URL")
		projectId, _ := strconv.Atoi(os.Getenv("PROJECT_ID"))
		imageName := os.Getenv("IMAGE_NAME")
		imageTag := os.Getenv("IMAGE_TAG")
		commitId := os.Getenv("COMMIT_ID")
		registryUrl := os.Getenv("REGISTRY_URL")
		hasDockerfile, _ := strconv.Atoi(os.Getenv("HAS_DOCKERFILE"))
		secret := os.Getenv("SECRET")
		dockerfileUrl := server + "/api/ci/build/builddockerfile/" + os.Getenv("PROJECT_ID") + "/" + os.Getenv("BUILD_ID")
		compilefileUrl := server + "/api/ci/build/compilefile/" + os.Getenv("PROJECT_ID") + "/" + os.Getenv("BUILD_ID")
		buildPath := os.Getenv("BUILD_PATH")
		dockerfilePath := os.Getenv("DOCKERFILE_PATH")
		buildType := os.Getenv("BUILD_TYPE")
		useAuth, _ := strconv.Atoi(os.Getenv("USE_AUTH"))

		buildContext := &buildcontext.BuildContext{idrsa, codeUrl, buildId, commitId, imageName, imageTag, registryUrl, hasDockerfile,
			secret, dockerfileUrl, compilefileUrl, buildPath, dockerfilePath, codeType, buildType, useAuth}
		script, err := buildContext.WriteScript()
		bs := &buildStatus{}
		bc, _ := json.Marshal(buildContext)
		bs.BuildId = buildId
		bs.ProjectId = projectId
		if err != nil {
			bs.Status = Fail
			bs.Message = "generate ci script fail, parameter: " + string(bc)

		} else {
			//write script file
			err = writeScriptFile(script)
			if err != nil {
				bs.Status = Fail
				bs.Message = "write ci script fail, parameter: " + string(bc)
			} else {
				// execute script file
				execScript(script, bs, string(bc))
			}
		}
		uploadUrl := fmt.Sprintf(server + "/api/ci/build/upload/%d/%d?secret=%s", projectId, buildId, secret)
		uploadLogFile(bs, uploadUrl, server, secret, codeType)
	}

}
package build

import (
    "imagebuilder/buildcontext"
    "imagebuilder/contect"
    "net/http"
    "fmt"
    "encoding/json"
    "os"
    "os/exec"
    "bytes"
    "strconv"
    "io"
    "mime/multipart"
    "io/ioutil"
    "time"
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
    ProjectId   int `json:"projectId"`
    BuildId     int `json:"buildId"`
    Status      string `json:"status"`
    Message     string `json:"message"`
}

func (bs *buildStatus)setBuildStatus(server string, secret string, timeout time.Duration) error {
    text, err := json.Marshal(bs)
    if err != nil {
        fmt.Println("marshal build status json error,", err)
        return err
    } else {
        body := bytes.NewBuffer([]byte(text))
        client := http.Client{
            Timeout: timeout,
        }
        _, err = client.Post(server + "/api/ci/build/status?secret=" + secret, "application/json;charset=utf-8", body)
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
    resp_body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    fmt.Println(string(resp_body))
    return nil
}

func Main() {
    // get basic settings from env
    server := os.Getenv("SERVER")
    buildId, _ := strconv.Atoi(os.Getenv("BUILD_ID"))
    idrsa := os.Getenv("IDRSA")
    codeUrl := os.Getenv("CODE_URL")
    projectId, _ := strconv.Atoi(os.Getenv("PROJECT_ID"))
    imageName := os.Getenv("IMAGE_NAME")
    imageTag := os.Getenv("IMAGE_TAG")
    commitId := os.Getenv("COMMIT_ID")
    registryUrl := os.Getenv("REGISTRY_URL")
    hasDockerfile,_ := strconv.Atoi(os.Getenv("HAS_DOCKERFILE"))
    secret := os.Getenv("SECRET")
    dockerfileUrl := server + "/api/ci/build/builddockerfile/" + os.Getenv("PROJECT_ID") + "/" + os.Getenv("BUILD_ID")
    buildPath := os.Getenv("BUILD_PATH")
    dockerfilePath := os.Getenv("DOCKERFILE_PATH")
	codeType := os.Getenv("TYPE")

//  init env for test
//	server = "10.2.86.175:8080"
//	buildId = 380
//	commitId = "rkasdf"
//	idrsa = "123456"
//	codeUrl = "svn://10.2.86.82/test123/trunk/ttttt"
//	projectId = 125
//	imageName = "test123/trunk/ttttt"
//	registryUrl = "10.11.150.76:5000"
//	hasDockerfile = 0
//	secret = "b3805b8e-3915-4228-a52d-6847550c6afa"
//	dockerfileUrl = server + "/api/ci/build/builddockerfile/" + "125/380"
//	codeType = "subversion"

    buildContext := &buildcontext.BuildContext{idrsa, codeUrl, commitId, imageName, imageTag, registryUrl, hasDockerfile, secret, dockerfileUrl, buildPath, dockerfilePath, codeType}

    script := buildContext.WriteScript()
    err := writeScriptFile(script)
    bs := &buildStatus{}
    bs.BuildId = buildId
    bs.ProjectId = projectId
    if err != nil {
        bs.Status = Fail
        bc, _ := json.Marshal(buildContext)
        bs.Message = "generate ci script fail, parameter: " + string(bc)
    } else {
        cmd := exec.Command("chmod", "755", ScriptFile)
        err := cmd.Run()
        if err != nil {
            fmt.Println("chmod 755 error")
            bs.Status = Fail
            bs.Message = "generate ci script fail"
        } else {
            executor := contect.Executor{}
            err = executor.Init()
            if err != nil {
                bs.Status = Fail
                bs.Message = "init log file error"
            } else {
                err := executor.Command("/bin/bash", ScriptFile)
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
    uploadUrl := fmt.Sprintf(server + "/api/ci/build/upload/%d/%d?secret=%s", projectId, buildId, secret)
    var uploadretry time.Duration = 0;
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
        err = bs.setBuildStatus(server, secret, timeout)
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
package buildcontext

import (
    "fmt"
    "imagebuilder/buildfile"
    "strings"
)

var LocalCodePath string = "/code"

// this struct stores all messages for build, parameters are provided by server
type BuildContext struct {
    Idrsa         string  `json:"idrsa"`

    CodeUrl       string  `json:"codeUrl"`

    CommitId      string  `json:commitId`

    ImageName     string  `json:"imageName"`

    ImageTag      string  `json:"imageTag"`

    RegistryUrl   string  `json:"registryUrl"`

    HasDockerfile int     `json:"hasDockerfle"`

    Secret        string  `json:"secret"`

    DockerfileUrl string  `json:"dockerfileUrl"`

    BuildPath     string  `json:"buildPath"`

    DockerfilePath string `json:"dockerfilePath"`

	CodeType	  string  `json:"type"`
}

func (context *BuildContext) WriteScript() (script string) {
    f := buildfile.New()
	f.WriteCmdSilent(fmt.Sprintf("if [ ! -d \"%s\" ]; then mkdir %s; fi", LocalCodePath, LocalCodePath))
	f.WriteCmdSilent(fmt.Sprintf("cd %s", LocalCodePath))
	if(strings.EqualFold(context.CodeType, "gitlab")) {
		if len(context.Idrsa) != 0 {
			f.WriteCmdSilent(fmt.Sprintf("echo -e '%s' > $HOME/.ssh/id_rsa", context.Idrsa))
			f.WriteCmdSilent("chmod 600 $HOME/.ssh/id_rsa")
		}
		if len(context.CodeUrl) > 0 {
			f.WriteCmd("git init")
			f.WriteCmd("git rev-parse --is-inside-work-tree # timeout=10")
			f.WriteCmd(fmt.Sprintf("git config remote.origin.url %s # timeout=10", context.CodeUrl))
			f.WriteCmd(fmt.Sprintf("git -c core.askpass=true fetch --tags --progress %s +refs/heads/*:refs/remotes/origin/*", context.CodeUrl))
			f.WriteCmd(fmt.Sprintf("git checkout -f %s", context.CommitId))
		}
	} else if(strings.EqualFold(context.CodeType, "subversion")) {
		if len(context.CodeUrl) > 0 {
			f.WriteCmd(fmt.Sprintf("svn --username '%s' --password '%s' --no-auth-cache checkout %s %s", context.CommitId, context.Idrsa, context.CodeUrl, context.ImageName))
		}
	}

    imageInfo := ""
    if len(context.RegistryUrl) > 0 {
        imageInfo = context.RegistryUrl + "/"
    }
    imageInfo = imageInfo + context.ImageName
    if len(context.ImageTag) > 0 {
        imageInfo = imageInfo + ":" + context.ImageTag
    } else {
        imageInfo = imageInfo + ":latest"
    }

    if context.HasDockerfile == 0 && len(context.DockerfileUrl) > 0 {
        f.WriteCmd(fmt.Sprintf("curl --connect-timeout 60 -o %s \"%s?secret=%s\"", LocalCodePath + "/Dockerfile", context.DockerfileUrl, context.Secret))
//        f.WriteCmdSilent(fmt.Sprintf("curl -o %s \"%s?secret=%s\"", LocalCodePath + "/Dockerfile", context.DockerfileUrl, context.Secret))
        context.DockerfilePath = LocalCodePath
        f.WriteCmd(fmt.Sprintf("docker build --pull -t %s %s", imageInfo, context.DockerfilePath))
    } else {
        context.BuildPath = LocalCodePath + context.BuildPath
        context.DockerfilePath = LocalCodePath + context.DockerfilePath
        f.WriteCmd(fmt.Sprintf("cd %s", context.BuildPath))
        f.WriteCmd(fmt.Sprintf("docker build --pull -f %s -t %s %s", context.DockerfilePath, imageInfo, context.BuildPath))
    }

    f.WriteCmd(fmt.Sprintf("docker push %s", imageInfo))

    // clean local
    // f.WriteCmd(fmt.Sprintf("docker rmi %s", imageInfo))

    // f.WriteCmd(fmt.Sprintf("rm -rf %s", LocalCodePath))
    return f.String()
}

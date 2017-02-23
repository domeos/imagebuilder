package test

import (
	"imagebuilder/build"
	"testing"
	"os"
)


func Test_Base_Image(t *testing.T)  {
	var envs = map[string]string{
		"SERVER" : "http://10.2.86.150:8080",
		"IMAGEID": "2",
		"IMAGENAME" : "test",
		"IMAGETAG" : "1.0",
		"SECRET" : "346ffead-ee10-45c6-a3f3-d914ac257a3e",
		"DOCKERFILE" : "00e721117fa43f4367dc9ed0ec6a5159",
		"REGISTRYURL" : "http://10.11.150.76:5000",
	}
	for key, val := range envs {
		os.Setenv(key, val)
	}

	build.RunOnType("BASEIMAGEBUILD")

	for key, _ := range envs {
		os.Unsetenv(key)
	}
}

func Test_Java_Build(t *testing.T)  {
	var envs = map[string]string{
		"SERVER" : "http://10.11.150.76:13986",
		"BUILD_ID": "109",
		"IDRSA" : "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQCO01zF+HogbD86XrFjblKPhVR1/+dGMBHzBwc3pL/zizs3Edg4\nA2JcGG6sb3CoCHJMyNckfbTftA/k8OwmrWkrr4SSeydCb4aAoHHCMfpyMtxAzPVC\nlAt+yjKrEfNwGCZqdLdXN4A/d4NLPVbNwyVaJgbk8YKchk78UutOFoQknQIDAQAB\nAoGAGapv1H+XarYpEpMrq2OK4JGkIORQqjM/Nn3/1Qb9G4XcqUPCqCYricM2ODR6\neSezaor45mzUkRKpfImy1ix5ZoP7lMPLx77dBRuib6i6u527t1UBAvEEGCYD9IPr\ndMcR6HYv1Rig0vgKCWYofdI2sgIvx01fRFEs99w65Inlq/0CQQDaQN4Qy11YSTa3\nsdH8UBpKyhAfef2puqt/UnsZyf3yxclWpNeS3au6H3i0+dgkbbzlMG9zQ/6lKJy8\nQujDJZcPAkEAp4bxk3PLS97XQKS9FHbL+Tbsu6WSNmU7rCtomQS3aUqkM7zpquWq\nyMe3qUEHCbf6nhLVokXL+YLdYLujaXMpkwJBAM3OE1kU253P1CgeJxvc0R4rMk7s\nMvWlD+jM90XXQn92YKgyYxGbtD6bRLCrVFTtog0gwkeYG3zUMhAYq/Kw9KMCQDSr\nHC/7a6LCwHG2WSuh3abQOcUU3M71LLmIPC4/aVpU+SK69cugwPy2rWss4oWPrd8c\nlMWbo/Ehz2+mDk4MwrkCQQCWiCf/9FUX8PmiSFk4aMcz+88LR6ahzpPeiBTryQ7J\nVlnPh3NWJbLEsGANBO2ztUy3+sthu1rdVIpl5FiaNKff\n-----END RSA PRIVATE KEY-----\n",
		"CODE_URL" : "git@code.sohuno.com:kairen/simple.git",
		"PROJECT_ID" : "42",
		"IMAGE_NAME" : "admin/simple-maven-test",
		"IMAGE_TAG" : "master_6767065",
		"COMMIT_ID" : "6767065a7669317c812059d1f18251e32b66c379",
		"REGISTRY_URL" : "10.11.150.76:5000",
		"HAS_DOCKERFILE" : "0",
		"SECRET" : "b233dc18-4b9a-4f14-bc0a-775722e18a1b",
		"BUILD_PATH" : "",
		"DOCKERFILE_PATH" : "",
		"BUILD_TYPE" : "java",
		"USE_AUTH" : "1",
	}
	for key, val := range envs {
		os.Setenv(key, val)
	}

	build.RunOnType("gitlab")

	for key, _ := range envs {
		os.Unsetenv(key)
	}
	
}

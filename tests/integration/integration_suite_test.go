package integration_test

import (
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

const RcloneRemoteBucket = "dev:io/s3dev1-ws"
const TestFolderRoot = "./test_files"
const TestFileName = "test.txt"
const TestFileSize = "21"

var LocalFilePath = path.Join(TestFolderRoot, TestFileName)

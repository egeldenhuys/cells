package integration_test

import (
	"fmt"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"os/exec"
	"path"
)

func runRclone(args []string) (string, error) {
	cmd := exec.Command("rclone", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func uploadFileExpectSuccess(filePath string, destFolder string) (folderUuid string) {
	testUuid := uuid.New().String()
	out, err := runRclone([]string{"copy", filePath, destFolder + "/" + testUuid})
	Expect(err).NotTo(HaveOccurred())
	Expect(out).To(Equal(""))

	return testUuid
}

var _ = Describe("Rclone functionality", func() {
	Describe("copying files", func() {
		Context("from the local filesystem to a remote bucket", func() {
			var testUuid string

			BeforeEach(func() {
				testUuid = uuid.New().String()
			})

			AfterEach(func() {
				out, err := runRclone([]string{"purge", RcloneRemoteBucket + "/" + testUuid + "/"})
				Expect(err).NotTo(HaveOccurred())
				if err != nil {
					GinkgoWriter.Println(out, err)
				}
			})

			It("should copy the file", func() {
				out, err := runRclone([]string{"copy", LocalFilePath, RcloneRemoteBucket + "/" + testUuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(""))

				out, err = runRclone([]string{"lsf", "--files-only", "-R", "--csv", "--format", "p", RcloneRemoteBucket + "/" + testUuid + "/"})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(fmt.Sprintf("%s\n", TestFileName)))
			})

			It("should copy the file with the original timestamp", func() {
				fi, err := os.Stat(LocalFilePath)
				Expect(err).NotTo(HaveOccurred(), "Failed to stat the test file")
				originalTimestamp := fi.ModTime()

				out, err := runRclone([]string{"copy", LocalFilePath, RcloneRemoteBucket + "/" + testUuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(""))

				out, err = runRclone([]string{"lsf", "--files-only", "-R", "--csv", "--format", "pt", RcloneRemoteBucket + "/" + testUuid + "/"})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(fmt.Sprintf("%s,%s\n", TestFileName, originalTimestamp.Format("2006-01-02 15:04:05"))))
			})
		})

		Context("within the same remote bucket", func() {
			var testUuid string
			var setupUuid string // Uploaded folder

			BeforeEach(func() {
				testUuid = uuid.New().String()
				setupUuid = uploadFileExpectSuccess(LocalFilePath, RcloneRemoteBucket)

				DeferCleanup(func() {
					out, err := runRclone([]string{"purge", RcloneRemoteBucket + "/" + setupUuid + "/"})
					Expect(err).NotTo(HaveOccurred())
					Expect(out).To(Equal(""))

					out, err = runRclone([]string{"purge", RcloneRemoteBucket + "/" + testUuid + "/"})
					Expect(err).NotTo(HaveOccurred())
					Expect(out).To(Equal(""))
				})
			})

			It("should copy the file", func() {
				out, err := runRclone([]string{"copy", RcloneRemoteBucket + "/" + setupUuid + TestFileName, RcloneRemoteBucket + "/" + testUuid})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(""))

				out, err = runRclone([]string{"lsf", "--files-only", "-R", "--csv", "--format", "pt", RcloneRemoteBucket + "/" + testUuid + "/"})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(fmt.Sprintf("%s\n", TestFileName)))
			})

			It("should copy the file with the original timestamp", func() {
				fi, err := os.Stat(LocalFilePath)
				Expect(err).NotTo(HaveOccurred(), "Failed to stat the test file")
				originalTimestamp := fi.ModTime()

				out, err := runRclone([]string{"lsf", "--files-only", "--csv", "--format", "pt", RcloneRemoteBucket + "/" + testUuid + "/"})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(fmt.Sprintf("%s,%s\n", TestFileName, originalTimestamp.Format("2006-01-02 15:04:05"))))
			})
		})

		Context("from a remote bucket to the local filesystem", func() {
			var testUuid string
			var setupUuid string // Uploaded folder

			BeforeEach(func() {
				setupUuid = uploadFileExpectSuccess(LocalFilePath, RcloneRemoteBucket)
				testUuid = uuid.New().String()

				DeferCleanup(func() {
					out, err := runRclone([]string{"purge", RcloneRemoteBucket + "/" + setupUuid + "/"})
					Expect(err).NotTo(HaveOccurred())
					Expect(out).To(Equal(""))
					if err != nil {
						GinkgoWriter.Println(out, err)
					}

					out, err = runRclone([]string{"purge", RcloneRemoteBucket + "/" + testUuid + "/'"})
					Expect(err).NotTo(HaveOccurred())
					Expect(out).To(Equal(""))
					if err != nil {
						GinkgoWriter.Println(out, err)
					}

					err = os.RemoveAll(path.Join(path.Dir(LocalFilePath), testUuid))
					if err != nil {
						GinkgoWriter.Println(err)
					}
				})
			})

			It("should copy the file", func() {
				out, err := runRclone([]string{"copy", RcloneRemoteBucket + "/" + setupUuid + TestFileName, path.Join(path.Dir(LocalFilePath), testUuid)})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(""))

				_, err = os.Stat(path.Join(path.Dir(LocalFilePath), testUuid, TestFileName))
				Expect(err).NotTo(HaveOccurred(), "Failed to stat the downloaded test file")
			})

			It("should copy the file with the original timestamp", func() {
				out, err := runRclone([]string{"copy", RcloneRemoteBucket + "/" + setupUuid + TestFileName, path.Join(path.Dir(LocalFilePath), testUuid)})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(""))

				OriginalFi, err := os.Stat(LocalFilePath)
				Expect(err).NotTo(HaveOccurred(), "Failed to stat the test file")

				downloadedFi, err := os.Stat(path.Join(path.Dir(LocalFilePath), testUuid, TestFileName))
				Expect(err).NotTo(HaveOccurred(), "Failed to stat the downloaded test file")

				Expect(downloadedFi.ModTime()).To(Equal(OriginalFi.ModTime()), "Downloaded file has different timestamp than the original")
			})
		})
	})
})

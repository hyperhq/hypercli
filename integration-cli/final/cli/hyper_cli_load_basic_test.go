package main

import (
	"time"
	"os"
	"fmt"
	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

/// test invalid url //////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidUrlProtocal(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	invalidURL := "ftp://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", invalidURL)
	c.Assert(output, checker.Equals, "Error response from daemon: Get " + invalidURL + ": unsupported protocol scheme \"ftp\"\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidUrlHost(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	invalidHost := "invalidhost"
	invalidURL := "http://" + invalidHost + "/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", invalidURL)
	c.Assert(output, checker.Equals, "Error response from daemon: Get " + invalidURL + ": dial tcp: lookup invalidhost: no such host\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidUrlPath(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/notexist.tar")
	c.Assert(output, checker.Equals, "Error response from daemon: Got HTTP status code >= 400: 403 Forbidden\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}


//test invalid ContentType and ContentLength///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidContentType(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/readme.txt")
	c.Assert(output, checker.Equals, "Error response from daemon: Download failed: image archive format should be tar, gzip, bzip, or xz\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidContentLengthTooLarge(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	const MAX_LENGTH = 4294967295
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/largefile.tar")
	c.Assert(output, checker.Contains, fmt.Sprintf("should be greater than zero and less than or equal to %v\n", MAX_LENGTH))
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

//test invalid content///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidContentLengthZero(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	const MAX_LENGTH = 4294967295
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/emptyfile.tar")
	c.Assert(output, checker.Equals, fmt.Sprintf("Error response from daemon: The size of the image archive file is 0, should be greater than zero and less than or equal to %v\n", MAX_LENGTH))
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidContentUnrelated(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/readme.tar")
	c.Assert(output, checker.Contains, "invalid argument\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidUntarFail(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/nottar.tar")
	c.Assert(output, checker.Contains, "Untar re-exec error: exit status 1: output: unexpected EOF\n")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestLoadFromInvalidContentIncomplete(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	deleteAllImages()
	url := "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-repositories.tgz"
	output, exitCode, err := dockerCmdWithError("load", "-i", url)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	deleteAllImages()


	//// load this image will be OK, but after delete this image, there is a residual image with <none> tag occur.
	//url = "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-manifest.tgz"
	//output, exitCode, err = dockerCmdWithError("load", "-i", url)
	//c.Assert(output, check.Not(checker.Contains), "has been loaded.")
	//c.Assert(exitCode, checker.Equals, 0)
	//c.Assert(err, checker.IsNil)
	//
	//images, _ = dockerCmd(c, "images", "hello-world")
	//c.Assert(images, checker.Contains, "hello-world")
	//
	//deleteAllImages()


	url = "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-layer.tgz"
	output, exitCode, err = dockerCmdWithError("load", "-i", url)
	c.Assert(output, checker.Contains, "json: no such file or directory")
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)

	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, check.Not(checker.Contains), "hello-world")

	deleteAllImages()
}

//test normal///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromPublicURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", publicURL)
	c.Assert(output, checker.Contains, "hello-world:latest(sha256:")
	c.Assert(output, checker.HasSuffix, "has been loaded.\n")
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromCompressedArchive(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	extAry := [...]string{"tar.gz", "tgz", "tar.bz2", "tar.xz"}

	for _, val := range extAry {
		publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld." + val
		output, exitCode, err := dockerCmdWithError("load", "-i", publicURL)
		c.Assert(output, checker.Contains, "hello-world:latest(sha256:")
		c.Assert(output, checker.HasSuffix, "has been loaded.\n")
		c.Assert(exitCode, checker.Equals, 0)
		c.Assert(err, checker.IsNil)

		time.Sleep(1 * time.Second)
	}
}

func (s *DockerSuite) TestLoadFromPublicURLWithQuiet(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	out, _, _ := dockerCmdWithStdoutStderr(c, "load", "-q", "-i", publicURL)
	c.Assert(out, check.Equals, "")

	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromPublicURLMultipeImage(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	multiImgURL := "http://image-tarball.s3.amazonaws.com/test/public/busybox_alpine.tar"
	dockerCmd(c, "load", "-i", multiImgURL)

	images, _ := dockerCmd(c, "images", "busybox")
	c.Assert(images, checker.Contains, "busybox")

	images, _ = dockerCmd(c, "images", "alpine")
	c.Assert(images, checker.Contains, "alpine")
}

func (s *DockerSuite) TestLoadFromBasicAuthURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	urlWithAuth := os.Getenv("URL_WITH_BASIC_AUTH")
	c.Assert(urlWithAuth, checker.NotNil)

	dockerCmd(c, "load", "-i", urlWithAuth)

	images, _ := dockerCmd(c, "images", "ubuntu")
	c.Assert(images, checker.Contains, "ubuntu")
}

func (s *DockerSuite) TestLoadFromAWSS3PreSignedURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	deleteAllImages()

	s3Region := "us-west-1"
	s3Bucket := "image-tarball"
	s3Key := "test/private/cirros.tar"
	preSignedUrl, err_ := generateS3PreSignedURL(s3Region, s3Bucket, s3Key)
	c.Assert(err_, checker.IsNil)
	time.Sleep(1 * time.Second)

	output, err := dockerCmd(c, "load", "-i", preSignedUrl)
	if err != 0 {
		fmt.Printf("preSignedUrl:[%v]\n", preSignedUrl)
		fmt.Printf("output:\n%v\n", output)
	}
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.Equals, 0)

	checkImage(c, true, "cirros")
}

//Prerequisite: update image balance to 1 in tenant collection of hypernetes in mongodb
//db.tenant.update({tenantid:"<tenant_id>"},{$set:{"resourceinfo.balance.images":2}})
func (s *DockerSuite) TestLoadFromPublicURLWithQuota(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	deleteAllImages()

	helloworldURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	multiImgURL := "http://image-tarball.s3.amazonaws.com/test/public/busybox_alpine.tar"
	ubuntuURL := "http://image-tarball.s3.amazonaws.com/test/public/ubuntu.tar.gz"
	exceedQuotaMsg := "Exceeded quota, please either delete images, or email support@hyper.sh to request increased quota"

	///// [init] /////
	// balance 2, images 0
	out, _ := dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 0")


	///// [step 1] load new hello-world image /////
	// balance 2 -> 1, image: 0 -> 1
	dockerCmd(c, "load", "-i", helloworldURL)
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 1")


	///// [step 2] load hello-world image again /////
	// balance 1 -> 1, image 1 -> 1
	output, exitCode, err := dockerCmdWithError("load", "-i", helloworldURL)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	checkImage(c, true, "hello-world")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 1")


	///// [step 3] load multiple image(busybox+alpine) /////
	// balance 1 -> 0, image 1 -> 2
	output, exitCode, err = dockerCmdWithError("load", "-i", multiImgURL)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)

	checkImage(c, true, "busybox")
	checkImage(c, false, "alpine")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 2")


	///// [step 4] load hello-world image again /////
	// balance 0 -> 0, image 2 -> 2
	output, exitCode, err = dockerCmdWithError("load", "-i", helloworldURL)
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)

	checkImage(c, true, "hello-world")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 2")


	///// [step 5] load new ubuntu image /////
	// balance 0 -> 0, image 2 -> 2
	output, exitCode, err = dockerCmdWithError("load", "-i", ubuntuURL)
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(err, checker.NotNil)

	checkImage(c, false, "ubuntu")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 2")


	///// [step 6] remove hello-world image /////
	// balance 0 -> 1, image 2 -> 1
	images, _ = dockerCmd(c, "rmi", "-f", "hello-world")
	c.Assert(images, checker.Contains, "Untagged: hello-world:latest")

	checkImage(c, false, "hello-world")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 1")


	///// [step 7] load new ubuntu image again /////
	//balance 1 -> 0, image 1 -> 2
	output, exitCode, err = dockerCmdWithError("load", "-i", ubuntuURL)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	checkImage(c, true, "ubuntu")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 2")


	///// [step 8] remove busybox and ubuntu image /////
	// balance 0 -> 2, image 2 -> 0
	images, _ = dockerCmd(c, "rmi", "-f", "busybox", "ubuntu:14.04")
	c.Assert(images, checker.Contains, "Untagged: busybox:latest")
	c.Assert(images, checker.Contains, "Untagged: ubuntu:14.04")

	checkImage(c, false, "busybox")
	checkImage(c, false, "ubuntu")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 0")


	///// [step 9] load multiple image(busybox+alpine) again /////
	// balance 2 -> 0, image 0 -> 2
	output, exitCode, err = dockerCmdWithError("load", "-i", multiImgURL)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	checkImage(c, true, "busybox")
	checkImage(c, true, "alpine")

	out, _ = dockerCmd(c, "info")
	c.Assert(out, checker.Contains, "Images: 2")
}

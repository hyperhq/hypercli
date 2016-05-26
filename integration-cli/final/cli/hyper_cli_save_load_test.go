package main

import (
	"time"
	"os"
	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

/// test invalid url //////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidUrlProtocal(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	invalidURL := "ftp://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", invalidURL)
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Equals, "Get " + invalidURL + ": unsupported protocol scheme \"ftp\"\n")
}

func (s *DockerSuite) TestLoadFromInvalidUrlHost(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	invalidHost := "invalidhost"
	invalidURL := "http://" + invalidHost + "/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", invalidURL)
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Equals, "Get " + invalidURL + ": dial tcp: lookup invalidhost: no such host\n")
}

func (s *DockerSuite) TestLoadFromInvalidUrlPath(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/notexist.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Equals, "Got HTTP status code >= 400: 403 Forbidden\n")
}


//test invalid ContentType and ContentLength///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidContentType(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/readme.txt")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Equals, "Download failed: image archive format should be tar, gzip, bzip, or xz\n")
}

func (s *DockerSuite) TestLoadFromInvalidContentLength(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/largefile.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "Download failed: image archive size is 2147491840, should be less than 2147483647\n")
}

//test invalid content///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromInvalidArchiveEmpty(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/emptyfile.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "invalid character 'i' looking for beginning of value")
}

func (s *DockerSuite) TestLoadFromInvalidContentUnrelated(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/readme.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "invalid character 'i' looking for beginning of value")
}

func (s *DockerSuite) TestLoadFromInvalidUntarFail(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	output, exitCode, err := dockerCmdWithError("load", "-i", "http://image-tarball.s3.amazonaws.com/test/public/nottar.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "Untar re-exec error: exit status 1: output: unexpected EOF")

}

func (s *DockerSuite) TestLoadFromInvalidContentIncomplete(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	deleteAllImages()
	url := "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-repositories.tgz"
	output, exitCode, err := dockerCmdWithError("load", "-i", url)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(output, checker.Contains, "has been loaded.")
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	deleteAllImages()
	url = "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-manifest.tgz"
	output, exitCode, err = dockerCmdWithError("load", "-i", url)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(output, check.Not(checker.Contains), "has been loaded.")
	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	deleteAllImages()
	url = "http://image-tarball.s3.amazonaws.com/test/public/helloworld-no-layer.tgz"
	output, exitCode, err = dockerCmdWithError("load", "-i", url)
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "json: no such file or directory")
	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, check.Not(checker.Contains), "hello-world")
}

//test normal///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromPublicURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	output, exitCode, err := dockerCmdWithError("load", "-i", publicURL)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(output, checker.Contains, "hello-world:latest(sha256:")
	c.Assert(output, checker.HasSuffix, "has been loaded.\n")

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
		c.Assert(err, checker.IsNil)
		c.Assert(exitCode, checker.Equals, 0)
		c.Assert(output, checker.Contains, "hello-world:latest(sha256:")
		c.Assert(output, checker.HasSuffix, "has been loaded.\n")
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

func (s *DockerSuite) TestLoadFromS3PreSignedURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	s3Region := "us-west-1"
	s3Bucket := "image-tarball"
	s3Key := "test/private/cirros.tar"
	preSignedUrl, err := generateS3PreSignedURL(s3Region, s3Bucket, s3Key)
	c.Assert(err, checker.IsNil)

	dockerCmd(c, "load", "-i", preSignedUrl)

	images, _ := dockerCmd(c, "images", "cirros")
	c.Assert(images, checker.Contains, "cirros")
}


//Prerequisite: update image balance to 1 in tenant collection of hypernetes in mongodb
//db.tenant.update({tenantid:"<tenant_id>"},{$set:{"resourceinfo.balance.images":2}})
func (s *DockerSuite) TestLoadFromPublicURLWithBalance(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	multiImgURL := "http://image-tarball.s3.amazonaws.com/test/public/busybox_alpine.tar"
	exceedQuotaMsg := "Exceeded quota, please either delete images, or email support@hyper.sh to request increased quota"
	s3Region := "us-west-1"
	s3Bucket := "image-tarball"
	s3Key := "test/private/cirros.tar"

	//balance 2 -> 1: load hello-world image(new)
	dockerCmd(c, "load", "-i", publicURL)
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 1 -> 1: load hello-world image again(existed)
	output, exitCode, err := dockerCmdWithError("load", "-i", publicURL)
	c.Assert(err, checker.IsNil)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 1 -> 0: load busybox alpine image(multiple image)
	output, exitCode, err = dockerCmdWithError("load", "-i", multiImgURL)
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)

	images, _ = dockerCmd(c, "images", "busybox")
	c.Assert(images, checker.Contains, "busybox")

	images, _ = dockerCmd(c, "images", "alpine")
	c.Assert(images, check.Not(checker.Contains), "alpine")

	//balance 0 -> 0: load hello-world image again(exist)
	output, exitCode, err = dockerCmdWithError("load", "-i", publicURL)
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)

	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 0 -> 0: load new cirros
	preSignedUrl, err := generateS3PreSignedURL(s3Region, s3Bucket, s3Key)
	c.Assert(err, checker.IsNil)
	output, exitCode, err = dockerCmdWithError("load", "-i", preSignedUrl)
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, exceedQuotaMsg)
	c.Assert(exitCode, checker.Equals, 1)
	images, _ = dockerCmd(c, "images", "cirros")
	c.Assert(images, check.Not(checker.Contains), "cirros")

	//balance 0 -> 1: remove hello-world image
	images, _ = dockerCmd(c, "rmi", "-f", "hello-world")
	c.Assert(images, checker.Contains, "Untagged: hello-world:latest")

	//balance 1 -> 0: load cirros image(new)
	output, exitCode, err = dockerCmdWithError("load", "-i", preSignedUrl)
	images, _ = dockerCmd(c, "images", "cirros")
	c.Assert(images, checker.Contains, "cirros")
}

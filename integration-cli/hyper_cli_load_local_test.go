package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

//test normal///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromLocalTarPipe(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	imagePath := fmt.Sprintf("%s/helloworld.tar", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	_, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(pathExist(imagePath), checker.Equals, true)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	//load via pipe
	catCmd := exec.Command("cat", imagePath)
	loadCmd := exec.Command(dockerBinary, "-H", os.Getenv("DOCKER_HOST"), "--config", os.Getenv("HYPER_CONFIG"), "load")

	catOut, err := catCmd.StdoutPipe()
	catCmd.Start()

	loadCmd.Stdin = catOut
	output, err := loadCmd.Output()
	c.Assert(string(output), checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)

	//check image
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromLocalTar(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/helloworld.tar"
	imagePath := fmt.Sprintf("%s/helloworld.tar", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(pathExist(imagePath), checker.Equals, true)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromLocalTarDelta(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	baseURL := "http://image-tarball.s3.amazonaws.com/test/public/debian_stretch-slim.tar.gz"
	basePath := fmt.Sprintf("%s/debian_stretch-slim.tar.gz", os.Getenv("IMAGE_DIR"))

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/nginx_stable.tar.gz"
	imagePath := fmt.Sprintf("%s/nginx_stable.tar.gz", os.Getenv("IMAGE_DIR"))

	//download base image tar
	wgetCmd := exec.Command("wget", "-cO", basePath, baseURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)
	c.Assert(pathExist(basePath), checker.Equals, true)

	wgetCmd = exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err = runCommandWithOutput(wgetCmd)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)
	c.Assert(pathExist(imagePath), checker.Equals, true)

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", basePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ := dockerCmd(c, "images", "debian:stretch-slim")
	c.Assert(images, checker.Contains, "debian")

	//check image
	images, _ = dockerCmd(c, "images", "nginx:stable")
	c.Assert(images, checker.Contains, "nginx")
}

func (s *DockerSuite) TestLoadFromLocalCompressedArchive(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	extAry := [...]string{"tar.gz", "tgz", "tar.bz2", "tar.xz"}

	//download image archive
	publicURL := ""
	imagePath := ""
	for _, val := range extAry {
		publicURL = "http://image-tarball.s3.amazonaws.com/test/public/helloworld." + val
		imagePath = fmt.Sprintf("%s/helloworld.%s", os.Getenv("IMAGE_DIR"), val)
		wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
		_, exitCode, err := runCommandWithOutput(wgetCmd)
		c.Assert(pathExist(imagePath), checker.Equals, true)
		c.Assert(err, checker.IsNil)
		c.Assert(exitCode, checker.Equals, 0)
	}

	//load image archive
	for _, val := range extAry {
		imagePath = fmt.Sprintf("%s/helloworld.%s", os.Getenv("IMAGE_DIR"), val)
		output, exitCode, err := dockerCmdWithError("load", "-i", imagePath)
		c.Assert(output, checker.Contains, "has been loaded.")
		c.Assert(err, checker.IsNil)
		c.Assert(exitCode, checker.Equals, 0)
		time.Sleep(1 * time.Second)
	}
}

func (s *DockerSuite) TestLoadFromLocalDocker(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	testImage := "hello-world:latest"

	//local docker pull image
	pullCmd := exec.Command("docker", "-H", os.Getenv("LOCAL_DOCKER_HOST"), "pull", testImage)
	output, exitCode, err := runCommandWithOutput(pullCmd)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//load image from local docker to hyper
	output, exitCode, err = dockerCmdWithError("load", "-l", testImage)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromLocalTarSize100MB(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/nginx_stable.tar"
	imagePath := fmt.Sprintf("%s/nginx_stable.tar", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)
	c.Assert(pathExist(imagePath), checker.Equals, true)

	//ensure nginx:stable not exist
	dockerCmdWithError("rmi", "nginx:stable")
	images, _ := dockerCmd(c, "images", "nginx:stable")
	c.Assert(images, checker.Not(checker.Contains), "nginx")

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ = dockerCmd(c, "images", "nginx:stable")
	c.Assert(images, checker.Contains, "nginx")
}

func (s *DockerSuite) TestLoadFromLocalTarSize600MB(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/jenkins.tar"
	imagePath := fmt.Sprintf("%s/jenkins.tar", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)
	c.Assert(pathExist(imagePath), checker.Equals, true)

	//ensure jenkins:latest not exist
	dockerCmdWithError("rmi", "jenkins:latest")
	images, _ := dockerCmd(c, "images", "jenkins:latest")
	c.Assert(images, checker.Not(checker.Contains), "jenkins")

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ = dockerCmd(c, "images", "jenkins:latest")
	c.Assert(images, checker.Contains, "jenkins")
}

func (s *DockerSuite) TestLoadFromLocalPullAndLoad(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/debian-8_5.tar.gz"
	imagePath := fmt.Sprintf("%s/debian-8_5.tar.gz", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)
	c.Assert(pathExist(imagePath), checker.Equals, true)

	//ensure debian:8.5 exist
	dockerCmdWithError("pull", "debian:8.5")
	images, _ := dockerCmd(c, "images", "debian:8.5")
	c.Assert(images, checker.Contains, "debian")

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ = dockerCmd(c, "images", "debian:8.5")
	c.Assert(images, checker.Contains, "debian")
}

//test abnormal///////////////////////////////////////////////////////////////////////////
func (s *DockerSuite) TestLoadFromLocalMultipeImage(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	multiImgURL := "http://image-tarball.s3.amazonaws.com/test/public/busybox_alpine.tar"
	imagePath := fmt.Sprintf("%s/busybox_alpine.tar", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, multiImgURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(pathExist(imagePath), checker.Equals, true)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "Loading multiple images from local is not supported")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Not(checker.Equals), 0)

	//ensure image not exist
	images, _ := dockerCmd(c, "images", "busybox")
	c.Assert(images, checker.Not(checker.Contains), "busybox")

	images, _ = dockerCmd(c, "images", "alpine")
	c.Assert(images, checker.Not(checker.Contains), "alpine")
}

func (s *DockerSuite) TestLoadFromLocalTarEmpty(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	//generate empty image tar
	imagePath := fmt.Sprintf("%s/empty.tar", os.Getenv("IMAGE_DIR"))
	os.OpenFile(imagePath, os.O_RDONLY|os.O_CREATE, 0666)
	f, err := os.OpenFile(imagePath, os.O_CREATE, 0600)
	c.Assert(err, checker.IsNil)
	f.Close()

	//load image tar
	output, exitCode, err := dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "manifest.json: no such file or directory")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Not(checker.Equals), 0)
}

func (s *DockerSuite) TestLoadFromLocalTarLegacy(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	publicURL := "http://image-tarball.s3.amazonaws.com/test/public/old/ubuntu_1.8.tar.gz"
	imagePath := fmt.Sprintf("%s/ubuntu_1.8.tar.gz", os.Getenv("IMAGE_DIR"))

	//download image tar
	wgetCmd := exec.Command("wget", "-cO", imagePath, publicURL)
	output, exitCode, err := runCommandWithOutput(wgetCmd)
	c.Assert(pathExist(imagePath), checker.Equals, true)
	c.Assert(exitCode, checker.Equals, 0)
	c.Assert(err, checker.IsNil)

	//load image tar
	output, exitCode, err = dockerCmdWithError("load", "-i", imagePath)
	c.Assert(output, checker.Contains, "manifest.json: no such file or directory")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Not(checker.Equals), 0)
}

/*
// TODO
//Prerequisite: update image balance to 1 in tenant collection of hypernetes in mongodb
//db.tenant.update({tenantid:"<tenant_id>"},{$set:{"resourceinfo.balance.images":2}})
func (s *DockerSuite) TestLoadFromLocalWithQuota(c *check.C) {
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
*/

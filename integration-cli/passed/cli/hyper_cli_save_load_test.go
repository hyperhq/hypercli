package main

import (
	//"encoding/json"
	//"fmt"
	//"io/ioutil"
	//"os"
	//"os/exec"
	//"path/filepath"
	//"reflect"
	//"regexp"
	//"sort"
	//"strings"
	"time"

	//"github.com/docker/distribution/digest"
	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

//// save a repo using gz compression and try to load it using stdout
//func (s *DockerSuite) TestSaveXzAndLoadRepoStdout(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	name := "test-save-xz-and-load-repo-stdout"
//	dockerCmd(c, "run", "--name", name, "busybox", "true")
//
//	repoName := "foobar-save-load-test-xz-gz"
//	out, _ := dockerCmd(c, "commit", name, repoName)
//
//	dockerCmd(c, "inspect", repoName)
//
//	repoTarball, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", repoName),
//		exec.Command("xz", "-c"),
//		exec.Command("gzip", "-c"))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save repo: %v %v", out, err))
//	deleteImages(repoName)
//
//	loadCmd := exec.Command(dockerBinary, "load")
//	loadCmd.Stdin = strings.NewReader(repoTarball)
//	out, _, err = runCommandWithOutput(loadCmd)
//	c.Assert(err, checker.NotNil, check.Commentf("expected error, but succeeded with no error and output: %v", out))
//
//	after, _, err := dockerCmdWithError("inspect", repoName)
//	c.Assert(err, checker.NotNil, check.Commentf("the repo should not exist: %v", after))
//}

//// save a repo using xz+gz compression and try to load it using stdout
//func (s *DockerSuite) TestSaveXzGzAndLoadRepoStdout(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	name := "test-save-xz-gz-and-load-repo-stdout"
//	dockerCmd(c, "run", "--name", name, "busybox", "true")
//
//	repoName := "foobar-save-load-test-xz-gz"
//	dockerCmd(c, "commit", name, repoName)
//
//	dockerCmd(c, "inspect", repoName)
//
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", repoName),
//		exec.Command("xz", "-c"),
//		exec.Command("gzip", "-c"))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save repo: %v %v", out, err))
//
//	deleteImages(repoName)
//
//	loadCmd := exec.Command(dockerBinary, "load")
//	loadCmd.Stdin = strings.NewReader(out)
//	out, _, err = runCommandWithOutput(loadCmd)
//	c.Assert(err, checker.NotNil, check.Commentf("expected error, but succeeded with no error and output: %v", out))
//
//	after, _, err := dockerCmdWithError("inspect", repoName)
//	c.Assert(err, checker.NotNil, check.Commentf("the repo should not exist: %v", after))
//}

//func (s *DockerSuite) TestSaveSingleTag(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	repoName := "foobar-save-single-tag-test"
//	dockerCmd(c, "tag", "busybox:latest", fmt.Sprintf("%v:latest", repoName))
//
//	out, _ := dockerCmd(c, "images", "-q", "--no-trunc", repoName)
//	cleanedImageID := strings.TrimSpace(out)
//
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", fmt.Sprintf("%v:latest", repoName)),
//		exec.Command("tar", "t"),
//		exec.Command("grep", "-E", fmt.Sprintf("(^repositories$|%v)", cleanedImageID)))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save repo with image ID and 'repositories' file: %s, %v", out, err))
//}

//func (s *DockerSuite) TestSaveCheckTimes(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	repoName := "busybox:latest"
//	out, _ := dockerCmd(c, "inspect", repoName)
//	data := []struct {
//		ID      string
//		Created time.Time
//	}{}
//	err := json.Unmarshal([]byte(out), &data)
//	c.Assert(err, checker.IsNil, check.Commentf("failed to marshal from %q: err %v", repoName, err))
//	c.Assert(len(data), checker.Not(checker.Equals), 0, check.Commentf("failed to marshal the data from %q", repoName))
//	tarTvTimeFormat := "2006-01-02 15:04"
//	out, _, err = runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", repoName),
//		exec.Command("tar", "tv"),
//		exec.Command("grep", "-E", fmt.Sprintf("%s %s", data[0].Created.Format(tarTvTimeFormat), digest.Digest(data[0].ID).Hex())))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save repo with image ID and 'repositories' file: %s, %v", out, err))
//}

//func (s *DockerSuite) TestSaveImageId(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	repoName := "foobar-save-image-id-test"
//	dockerCmd(c, "tag", "emptyfs:latest", fmt.Sprintf("%v:latest", repoName))
//
//	out, _ := dockerCmd(c, "images", "-q", "--no-trunc", repoName)
//	cleanedLongImageID := strings.TrimPrefix(strings.TrimSpace(out), "sha256:")
//
//	out, _ = dockerCmd(c, "images", "-q", repoName)
//	cleanedShortImageID := strings.TrimSpace(out)
//
//	// Make sure IDs are not empty
//	c.Assert(cleanedLongImageID, checker.Not(check.Equals), "", check.Commentf("Id should not be empty."))
//	c.Assert(cleanedShortImageID, checker.Not(check.Equals), "", check.Commentf("Id should not be empty."))
//
//	saveCmd := exec.Command(dockerBinary, "save", cleanedShortImageID)
//	tarCmd := exec.Command("tar", "t")
//
//	var err error
//	tarCmd.Stdin, err = saveCmd.StdoutPipe()
//	c.Assert(err, checker.IsNil, check.Commentf("cannot set stdout pipe for tar: %v", err))
//	grepCmd := exec.Command("grep", cleanedLongImageID)
//	grepCmd.Stdin, err = tarCmd.StdoutPipe()
//	c.Assert(err, checker.IsNil, check.Commentf("cannot set stdout pipe for grep: %v", err))
//
//	c.Assert(tarCmd.Start(), checker.IsNil, check.Commentf("tar failed with error: %v", err))
//	c.Assert(saveCmd.Start(), checker.IsNil, check.Commentf("docker save failed with error: %v", err))
//	defer func() {
//		saveCmd.Wait()
//		tarCmd.Wait()
//		dockerCmd(c, "rmi", repoName)
//	}()
//
//	out, _, err = runCommandWithOutput(grepCmd)
//
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save repo with image ID: %s, %v", out, err))
//}

//// save a repo and try to load it using flags
//func (s *DockerSuite) TestSaveAndLoadRepoFlags(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	name := "test-save-and-load-repo-flags"
//	dockerCmd(c, "run", "--name", name, "busybox", "true")
//
//	repoName := "foobar-save-load-test"
//
//	deleteImages(repoName)
//	dockerCmd(c, "commit", name, repoName)
//
//	before, _ := dockerCmd(c, "inspect", repoName)
//
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", repoName),
//		exec.Command(dockerBinary, "load"))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save and load repo: %s, %v", out, err))
//
//	after, _ := dockerCmd(c, "inspect", repoName)
//	c.Assert(before, checker.Equals, after, check.Commentf("inspect is not the same after a save / load"))
//}

//func (s *DockerSuite) TestSaveMultipleNames(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	repoName := "foobar-save-multi-name-test"
//
//	// Make one image
//	dockerCmd(c, "tag", "emptyfs:latest", fmt.Sprintf("%v-one:latest", repoName))
//
//	// Make two images
//	dockerCmd(c, "tag", "emptyfs:latest", fmt.Sprintf("%v-two:latest", repoName))
//
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", fmt.Sprintf("%v-one", repoName), fmt.Sprintf("%v-two:latest", repoName)),
//		exec.Command("tar", "xO", "repositories"),
//		exec.Command("grep", "-q", "-E", "(-one|-two)"),
//	)
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save multiple repos: %s, %v", out, err))
//}

//func (s *DockerSuite) TestSaveRepoWithMultipleImages(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	makeImage := func(from string, tag string) string {
//		var (
//			out string
//		)
//		out, _ = dockerCmd(c, "run", "-d", from, "true")
//		cleanedContainerID := strings.TrimSpace(out)
//
//		out, _ = dockerCmd(c, "commit", cleanedContainerID, tag)
//		imageID := strings.TrimSpace(out)
//		return imageID
//	}
//
//	repoName := "foobar-save-multi-images-test"
//	tagFoo := repoName + ":foo"
//	tagBar := repoName + ":bar"
//
//	idFoo := makeImage("busybox:latest", tagFoo)
//	idBar := makeImage("busybox:latest", tagBar)
//
//	deleteImages(repoName)
//
//	// create the archive
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", repoName, "busybox:latest"),
//		exec.Command("tar", "t"))
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save multiple images: %s, %v", out, err))
//
//	lines := strings.Split(strings.TrimSpace(out), "\n")
//	var actual []string
//	for _, l := range lines {
//		if regexp.MustCompile("^[a-f0-9]{64}\\.json$").Match([]byte(l)) {
//			actual = append(actual, strings.TrimSuffix(l, ".json"))
//		}
//	}
//
//	// make the list of expected layers
//	out = inspectField(c, "busybox:latest", "Id")
//	expected := []string{strings.TrimSpace(out), idFoo, idBar}
//
//	// prefixes are not in tar
//	for i := range expected {
//		expected[i] = digest.Digest(expected[i]).Hex()
//	}
//
//	sort.Strings(actual)
//	sort.Strings(expected)
//	c.Assert(actual, checker.DeepEquals, expected, check.Commentf("archive does not contains the right layers: got %v, expected %v, output: %q", actual, expected, out))
//}

//// Issue #6722 #5892 ensure directories are included in changes
//func (s *DockerSuite) TestSaveDirectoryPermissions(c *check.C) {
//	testRequires(c, DaemonIsLinux)
//	layerEntries := []string{"opt/", "opt/a/", "opt/a/b/", "opt/a/b/c"}
//	layerEntriesAUFS := []string{"./", ".wh..wh.aufs", ".wh..wh.orph/", ".wh..wh.plnk/", "opt/", "opt/a/", "opt/a/b/", "opt/a/b/c"}
//
//	name := "save-directory-permissions"
//	tmpDir, err := ioutil.TempDir("", "save-layers-with-directories")
//	c.Assert(err, checker.IsNil, check.Commentf("failed to create temporary directory: %s", err))
//	extractionDirectory := filepath.Join(tmpDir, "image-extraction-dir")
//	os.Mkdir(extractionDirectory, 0777)
//
//	defer os.RemoveAll(tmpDir)
//	_, err = buildImage(name,
//		`FROM busybox
//	RUN adduser -D user && mkdir -p /opt/a/b && chown -R user:user /opt/a
//	RUN touch /opt/a/b/c && chown user:user /opt/a/b/c`,
//		true)
//	c.Assert(err, checker.IsNil, check.Commentf("%v", err))
//
//	out, _, err := runCommandPipelineWithOutput(
//		exec.Command(dockerBinary, "save", name),
//		exec.Command("tar", "-xf", "-", "-C", extractionDirectory),
//	)
//	c.Assert(err, checker.IsNil, check.Commentf("failed to save and extract image: %s", out))
//
//	dirs, err := ioutil.ReadDir(extractionDirectory)
//	c.Assert(err, checker.IsNil, check.Commentf("failed to get a listing of the layer directories: %s", err))
//
//	found := false
//	for _, entry := range dirs {
//		var entriesSansDev []string
//		if entry.IsDir() {
//			layerPath := filepath.Join(extractionDirectory, entry.Name(), "layer.tar")
//
//			f, err := os.Open(layerPath)
//			c.Assert(err, checker.IsNil, check.Commentf("failed to open %s: %s", layerPath, err))
//
//			entries, err := listTar(f)
//			for _, e := range entries {
//				if !strings.Contains(e, "dev/") {
//					entriesSansDev = append(entriesSansDev, e)
//				}
//			}
//			c.Assert(err, checker.IsNil, check.Commentf("encountered error while listing tar entries: %s", err))
//
//			if reflect.DeepEqual(entriesSansDev, layerEntries) || reflect.DeepEqual(entriesSansDev, layerEntriesAUFS) {
//				found = true
//				break
//			}
//		}
//	}
//
//	c.Assert(found, checker.Equals, true, check.Commentf("failed to find the layer with the right content listing"))
//
//}


//// Test loading a weird image where one of the layers is of zero size.
//// The layer.tar file is actually zero bytes, no padding or anything else.
//// See issue: 18170
//func (s *DockerSuite) TestLoadZeroSizeLayer(c *check.C) {
//	printTestCaseName(); defer printTestDuration(time.Now())
//	testRequires(c, DaemonIsLinux)
//	dockerCmd(c, "load", "-i", "http://hyper-upload.s3.amazonaws.com/image_tarball/test/public/emptyLayer.tar")
//}

func (s *DockerSuite) TestLoadFromInvalidUrlProtocal(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "tcp://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "unsupported protocol scheme \"tcp\"")
}

func (s *DockerSuite) TestLoadFromInvalidUrlPath(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	output, exitCode, err := dockerCmdWithError("load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/notexist.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(exitCode, checker.Equals, 1)
	c.Assert(output, checker.Contains, "403 Forbidden")
}


func (s *DockerSuite) TestLoadFromPublicURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	dockerCmd(c, "load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")

	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromPublicURLWithQuiet(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	out, _, _ := dockerCmdWithStdoutStderr(c, "load", "-q", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")
	c.Assert(out, check.Equals, "")

	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}

func (s *DockerSuite) TestLoadFromPublicURLMultipeImage(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	dockerCmd(c, "load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/busybox_alpine.tar")

	images, _ := dockerCmd(c, "images", "busybox")
	c.Assert(images, checker.Contains, "busybox")

	images, _ = dockerCmd(c, "images", "alpine")
	c.Assert(images, checker.Contains, "alpine")
}

func (s *DockerSuite) TestLoadFromBasicAuthURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	dockerCmd(c, "load", "-i", "http://hyper:aaa123aa@test.hyper.sh/docker.tar")

	images, _ := dockerCmd(c, "images", "docker")
	c.Assert(images, checker.Contains, "docker")
}

func (s *DockerSuite) TestLoadFromS3PreSignedURL(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	preSignedUrl, err := generateS3PreSignedURL("ap-northeast-1","hyper-upload","image_tarball/test/private/cirros.tar")
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

	//balance 2 -> 1: load hello-world image(new)
	dockerCmd(c, "load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 1 -> 1: load hello-world image again(existed)
	output, exitCode, err := dockerCmdWithError("load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")
	c.Assert(err, checker.IsNil)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(exitCode, checker.Equals, 0)
	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 1 -> 0: load busybox alpine image(multiple image)
	output, exitCode, err = dockerCmdWithError("load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/busybox_alpine.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(output, checker.Contains, "no balance available")
	c.Assert(exitCode, checker.Equals, 1)

	images, _ = dockerCmd(c, "images", "busybox")
	c.Assert(images, checker.Contains, "busybox")

	images, _ = dockerCmd(c, "images", "alpine")
	c.Assert(images, check.Not(checker.Contains), "alpine")

	//balance 0 -> 0: load hello-world image again(exist)
	output, exitCode, err = dockerCmdWithError("load", "-i", "https://hyper-upload.s3.amazonaws.com/image_tarball/test/public/helloworld.tar")
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, "no balance available")
	c.Assert(exitCode, checker.Equals, 1)

	images, _ = dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")

	//balance 0 -> 0: load cirros(not exist)
	preSignedUrl, err := generateS3PreSignedURL("ap-northeast-1","hyper-upload","image_tarball/test/private/cirros.tar")
	c.Assert(err, checker.IsNil)
	output, exitCode, err = dockerCmdWithError("load", "-i", preSignedUrl)
	c.Assert(err, checker.NotNil)
	c.Assert(output, checker.Contains, "no balance available")
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
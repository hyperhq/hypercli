Integration test for hyper cli
==================================

> functional test for hyper cli  
> use apirouter service on packet(dev env) as backend

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [Project status](#project-status)
	- [cli test case](#cli-test-case)
	- [extra](#extra)
	- [skip](#skip)
- [Command list](#command-list)
	- [hyper only](#hyper-only)
	- [both](#both)
	- [docker only](#docker-only)
- [Prepare](#prepare)
	- [clone hypercli repo](#clone-hypercli-repo)
	- [build docker image](#build-docker-image)
	- [make hyper cli in container](#make-hyper-cli-in-container)
	- [common info in container](#common-info-in-container)
- [Run test case](#run-test-case)
	- [enter container](#enter-container)
	- [run test in container](#run-test-in-container)
		- [(optional)test connection to apirouter service](#optionaltest-connection-to-apirouter-service)
		- [prepare test case](#prepare-test-case)
		- [adjust test case code](#adjust-test-case-code)
		- [start test](#start-test)
- [Check test result](#check-test-result)
	- [if test case passed](#if-test-case-passed)
	- [if find issues](#if-find-issues)
- [After issues fixed](#after-issues-fixed)

<!-- /TOC -->

# Project status

## cli test case

- [ ] cli_attach_test
- [ ] cli_attach_unix_test
- [ ] cli_config_test
- [ ] cli_create_test
- [ ] cli_exec_test
- [ ] cli_exec_unix_test
- [ ] cli_help_test
- [ ] cli_history_test
- [ ] cli_images_test
- [ ] cli_info_test
- [ ] cli_inspect_experimental_test
- [ ] cli_inspect_test
- [ ] cli_links_test
- [ ] cli_links_unix_test
- [ ] cli_login_test
- [ ] cli_logs_test
- [ ] cli_nat_test
- [ ] cli_netmode_test
- [ ] cli_network_unix_test
- [ ] cli_oom_killed_test
- [ ] cli_port_test
- [ ] cli_proxy_test
- [ ] cli_ps_test
- [ ] cli_rename_test
- [ ] cli_restart_test
- [ ] cli_rmi_test
- [ ] cli_rm_test
- [ ] cli_run_test
- [ ] cli_run_unix_test
- [ ] cli_search_test
- [ ] cli_sni_test
- [ ] cli_start_test
- [ ] cli_start_volume_driver_unix_test
- [ ] cli_stats_test
- [x] cli_version_test
- [ ] cli_volume_driver_compat_unix_test
- [ ] cli_volume_test


## extra

[Extra Test Case](EXTRA_TEST.md)


## skip

> not support build, commit, push, tag

- [ ] cli_authz_unix_test
- [ ] cli_build_test
- [ ] cli_build_unix_test
- [ ] cli_by_digest_test
- [ ] cli_commit_test
- [ ] cli_cp_from_container_test
- [ ] cli_cp_test
- [ ] cli_cp_to_container_test
- [ ] cli_cp_utils
- [ ] cli_daemon_test
- [ ] cli_diff_test
- [ ] cli_events_test
- [ ] cli_events_unix_test
- [ ] cli_experimental_test
- [ ] cli_export_import_test
- [ ] cli_external_graphdriver_unix_test
- [ ] cli_import_test
- [ ] cli_pause_test
- [ ] cli_pull_local_test
- [ ] cli_pull_trusted_test
- [ ] cli_push_test
- [ ] cli_save_load_test
- [ ] cli_save_load_unix_test
- [ ] cli_tag_test
- [ ] cli_top_test
- [ ] cli_update_unix_test
- [ ] cli_v2_only_test
- [ ] cli_wait_test

# Command list

| hyper only | both | docker only |
| --- | --- | --- |
| 3 | 25 | 17 |

## hyper only
```
config	fip	snapshot
```

## both
```
attach	create	exec	history	images
info	inspect	kill	login	logout
logs	port	ps	pull	rename
restart	rm	rmi	run	search
start	stats	stop	version	volume
```

## docker only

> not support for hyper currently

```
build	commit	cp	diff	events
export	import	load	network	pause
push	save	tag	top	unpause
update	wait			
```



# Prepare

## clone hypercli repo
```
$ git clone https://github.com/hyperhq/hypercli.git -b integration-test
```

## build docker image

> build docker image in host OS  
> Use `CentOS` as test env

```
// run in dir hypercli/integration-cli on host os
$ ./util.sh build
```

## make hyper cli in container

> build hyper cli binary from source code

```
// run in dir hypercli/integration-cli on host os
$ ./util.sh make
```

## common info in container

- work dir        : `/go/src/github.com/hyperhq/hypercli/integration-cli`
- hyper config    : `/root/.hyper/config.json`
- hyper cli binary: `/usr/bin/hyper` -> `/go/src/github.com/hyperhq/hypercli/hyper/hyper`
- hyper cli alias : `hypercli` => `hyper -H ${DOCKER_HOST}`
- test case dir   : `/go/src/github.com/hyperhq/hypercli/integration-cli`
```
integration-cli
├── skip      => test cases to be ignored
├── todo      => test cases to be tested
├── issue     => test cases have issue/bug
└── passed    => test cases have passed
```


# Run test case

## enter container

> update `ACCESS_KEY` and `SECRET_KEY` in `integration-cli/util.conf`
```
// run in dir hypercli/integration-cli on host os
$ ./util.sh enter
```

## run test in container

### (optional)test connection to apirouter service
```
// run in any dir in container
hypercli version
hypercli pull busybox
hypercli images
```

### prepare test case

- **test new case**: move test case from `integration-cli/todo` to `integration-cli` 
- **test issue case after fixed**: move test case from `integration-cli/issue` to `integration-cli` 

### adjust test case code

> add `printTestCaseName(); defer printTestDuration(time.Now())` in function start with `Test`
> hyper cli source will be mapped in to the container, so the test case code can be modified out of container

```
//example:
func (s *DockerSuite) TestVersionEnsureSucceeds(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())    <<<<<<======
	out, _ := dockerCmd(c, "version")

//test result will be output like:
[2016-04-26 03:21:52] github.com/hyperhq/hypercli/integration-cli.(*DockerSuite).TestVersionEnsureSucceeds - 1.952121 sec
```

### start test

```
// run in dir hypercli/integration-cli in container
$ go test
```

# Check test result

> Check the `passed` number of `go test` result

```
...
INFO: Testing against a remote daemon
...
OK: ? passed, ? skipped
PASS
ok      github.com/hyperhq/hypercli/integration-cli    ?s
```

## if test case passed

- move the test case to `integration-cli/passed` dir
- continue next test case

## if find issues

- move the test case to `integration-cli/issue` dir
- please create a new issue here: https://github.com/hyperhq/hypercli/issues
- continue next test case

# After issues fixed

- move the test case from `integration-cli/issue` to `integration-cli` dir
- go to [start test](#start-test)

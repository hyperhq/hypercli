Integration test for hyper cli
==================================

> functional test for hyper cli  
> use apirouter service on packet(dev env) as backend
> skip test-case for daemon
> skip test-case for build/commit/push/tag
> skip test-case fro save/load, import/export

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [Project status](#project-status)
	- [cli test case](#cli-test-case)
	- [extra](#extra)
	- [skip](#skip)
- [Prepare](#prepare)
	- [clone hypercli repo](#clone-hypercli-repo)
	- [build docker image](#build-docker-image)
	- [build hyper cli in container](#build-hyper-cli-in-container)
	- [common info in container](#common-info-in-container)
- [Run test case](#run-test-case)
	- [enter container](#enter-container)
	- [run test in container](#run-test-in-container)
		- [(optional)test connection to apirouter service](#optionaltest-connection-to-apirouter-service)
		- [prepare test case](#prepare-test-case)
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


# Prepare

## clone hypercli repo
```
git clone https://github.com/hyperhq/hypercli.git -b integration-test 
cd hypercli
```

## build docker image

> Use `CentOS` as test env

```
docker build -t hyperhq/hypercli -f Dockerfile.centos .
```

## build hyper cli in container

> run the following command in root of hypercli repo

```
docker run -it --rm -v $(pwd):/go/src/github.com/hyperhq/hypercli hyperhq/hypercli ./build-hyperserve-client.sh
```

## common info in container

- hyper config: `/root/.hyper/config.json`
- work dir: `/go/src/github.com/hyperhq/hypercli`
- test case dir: `/go/src/github.com/hyperhq/hypercli/integration-cli`
```
integration-cli
├── todo      => test cases to be tested
├── issue     => test cases have issue/bug
└── passed    => test cases have passed
```
- hyper cli binary:  
```
ll /usr/bin/hyper 
  lrwxrwxrwx 1 root root 47 Apr 21 08:59 /usr/bin/hyper -> /go/src/github.com/hyperhq/hypercli/hyper/hyper
```

# Run test case

## enter container

> replace `ACCESS_KEY` and `SECRET_KEY` with the real value
> run the following command in root of hypercli repo

```
docker run -it --rm \
   -e DOCKER_HOST=tcp://147.75.195.39:6443 \
   -e ACCESS_KEY=RE5xxxxxxxxxxxxxxxxxxxxxxxxxBP \
   -e SECRET_KEY=J7Kxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxvy \
   -v $(pwd):/go/src/github.com/hyperhq/hypercli \
   hyperhq/hypercli bash
```

## run test in container

### (optional)test connection to apirouter service
```
hyper version
hyper pull busybox
hyper images
```

### prepare test case

- **test new case**: move test case from `integration-cli/todo` to `integration-cli` 
- **test issue case after fixed**: move test case from `integration-cli/issue` to `integration-cli` 

### start test

> run test in `/go/src/github.com/hyperhq/hypercli/integration-cli` dir

```
go test
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

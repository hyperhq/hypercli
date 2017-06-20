Integration test for hyper cli
==================================

> functional test for hyper cli
> use apirouter service on packet(dev env) as backend

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [Project status](#project-status)
	- [cli test case](#cli-test-case)
	- [api test case](#api-test-case)
	- [extra](#extra)
	- [skip](#skip)
- [Command list](#command-list)
	- [hyper only](#hyper-only)
	- [both](#both)
	- [docker only](#docker-only)
- [Run test case](#run-test-case)
	- [clone hypercli repo](#clone-hypercli-repo)
	- [run test in docker container](#run-test-in-docker-container)
		- [build docker image](#build-docker-image)
		- [make hyper cli](#make-hyper-cli)
		- [enter container](#enter-container)
		- [run test in container](#run-test-in-container)
	- [Run test on localhost](#run-test-on-localhost)
		- [prepare](#prepare)
		- [run test case](#run-test-case)

<!-- /TOC -->

# Project status

## cli test case

- [ ] cli_attach_test
- [ ] cli_attach_unix_test
- [x] cli_config_test
- [x] cli_create_test
- [x] cli_exec_test
- [x] cli_exec_unix_test
- [x] cli_help_test
- [x] cli_history_test
- [x] cli_images_test
- [x] cli_info_test
- [x] cli_inspect_experimental_test
- [x] cli_inspect_test
- [x] cli_kill_test
- [ ] cli_links_test
- [ ] cli_links_unix_test
- [x] cli_login_test
- [x] cli_logs_test
- [x] cli_port_test
- [x] cli_ps_test
- [x] cli_pull_test
- [x] cli_rename_test
- [ ] cli_restart_test
- [x] cli_rm_test
- [x] cli_rmi_test
- [ ] cli_run_test
- [ ] cli_run_unix_test
- [x] cli_search_test
- [x] cli_start_test
- [ ] cli_stats_test
- [x] cli_version_test
- [x] cli_volume_test


## api test case

- [ ] api_attach_test
- [ ] api_containers_test
- [x] api_create_test
- [x] api_exec_test
- [x] api_exec_resize_test
- [x] api_images_test
- [x] api_info_test
- [x] api_inspect_test
- [x] api_logs_test
- [x] api_stats_test
- [x] api_snapshots_test
- [x] api_version_test
- [x] api_volumes_test

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
- [ ] cli_nat_test
- [ ] cli_netmode_test
- [ ] cli_network_unix_test
- [ ] cli_oom_killed_test
- [ ] cli_pause_test
- [ ] cli_proxy_test
- [ ] cli_pull_local_test
- [ ] cli_pull_trusted_test
- [ ] cli_push_test
- [ ] cli_save_load_test
- [ ] cli_save_load_unix_test
- [ ] cli_sni_test
- [ ] cli_start_volume_driver_unix_test
- [ ] cli_tag_test
- [ ] cli_top_test
- [ ] cli_update_unix_test
- [ ] cli_v2_only_test
- [ ] cli_volume_driver_compat_unix_test
- [ ] cli_wait_test


# Command list

| hyper only | both | docker only |
| --- | --- | --- |
| 3 | 25 | 17 |

## hyper only
```
config  fip     snapshot
```

## both
```
attach	create	exec	history	images
info    inspect	kill    login   logout
logs    port    ps      pull    rename
restart rm      rmi     run     search
start   stats   stop    version volume
```

## docker only

> not support for hyper currently

```
build   commit  cp      diff    events
export  import  load    network pause
push    save    tag     top     unpause
update  wait
```

# Run test case

## clone hypercli repo
```
$ git clone https://github.com/hyperhq/hypercli.git
```

## run test in docker container

### build docker image

> build docker image in host OS
> Use `CentOS` as test env

```
// run in dir hypercli/integration-cli on host os
$ ./util.sh build
```

### make hyper cli

> build hyper cli binary from source code

```
// run in dir hypercli/integration-cli on host os
$ ./util.sh make
```

### enter container

> update `ACCESS_KEY` and `SECRET_KEY` in `integration-cli/util.conf`

```
// run in dir hypercli/integration-cli on host os
$ ./util.sh enter
```

### run test in container
```
$ ./util.sh test all
```

## Run test on localhost

### prepare

```
// ensure hyperhq and docker dir
mkdir -p $GOPATH/src/github.com/{hyperhq,docker}

// clone and build hypercli
cd $GOPATH/src/github.com/hyperhq
git clone git@github.com:hyperhq/hypercli.git
cd hypercli
./build.sh

// copy hyper binary to /usr/bin/hyper
sudo cp hyper/hyper /usr/bin/hyper

// create link
cd $GOPATH/src/github.com/docker
ln -s ../hyperhq/hypercli docker

// generate util.conf
$ cd $GOPATH/src/github.com/hyperhq/hypercli
$ git checkout integration-test
$ cd integration-cli
$ ./util.sh

// config util.conf
$ vi util.conf
ACCESS_KEY="<hyper access key>"
SECRET_KEY="<hyper secret key>"
```

### run test case

```
// run all test cases
$ ./util.sh test all

// run test case with timeout
$ ./util.sh test all -timout 20m

// run specified test case
$ ./util.sh test -check.f ^TestCliLoadFromLocalTar$

// run test cases start with specified prefix
$ ./util.sh test -check.f TestCliLoadFromLocalTar

// combined use
$ ./util.sh test -check.f 'TestCliLoadFromLocalTarEmpty|TestCliLoadFromLocalPullAndLoad' -timeout 20m
```

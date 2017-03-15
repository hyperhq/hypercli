package container

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	timetypes "github.com/docker/engine-api/types/time"
	"github.com/hyperhq/hypercli/api/server/httputils"
	"github.com/hyperhq/hypercli/daemon"
	derr "github.com/hyperhq/hypercli/errors"
	"github.com/hyperhq/hypercli/pkg/ioutils"
	"github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/pkg/term"
	"github.com/hyperhq/hypercli/runconfig"
	"github.com/hyperhq/hypercli/utils"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

func (s *containerRouter) getContainersJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	config := &daemon.ContainersConfig{
		All:     httputils.BoolValue(r, "all"),
		Size:    httputils.BoolValue(r, "size"),
		Since:   r.Form.Get("since"),
		Before:  r.Form.Get("before"),
		Filters: r.Form.Get("filters"),
	}

	if tmpLimit := r.Form.Get("limit"); tmpLimit != "" {
		limit, err := strconv.Atoi(tmpLimit)
		if err != nil {
			return err
		}
		config.Limit = limit
	}

	containers, err := s.backend.Containers(config)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, containers)
}

func (s *containerRouter) getContainersStats(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	stream := httputils.BoolValueOrDefault(r, "stream", true)
	var out io.Writer
	if !stream {
		w.Header().Set("Content-Type", "application/json")
		out = w
	} else {
		wf := ioutils.NewWriteFlusher(w)
		out = wf
		defer wf.Close()
	}

	var closeNotifier <-chan bool
	if notifier, ok := w.(http.CloseNotifier); ok {
		closeNotifier = notifier.CloseNotify()
	}

	config := &daemon.ContainerStatsConfig{
		Stream:    stream,
		OutStream: out,
		Stop:      closeNotifier,
		Version:   httputils.VersionFromContext(ctx),
	}

	return s.backend.ContainerStats(vars["name"], config)
}

func (s *containerRouter) getContainersLogs(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	// Args are validated before the stream starts because when it starts we're
	// sending HTTP 200 by writing an empty chunk of data to tell the client that
	// daemon is going to stream. By sending this initial HTTP 200 we can't report
	// any error after the stream starts (i.e. container not found, wrong parameters)
	// with the appropriate status code.
	stdout, stderr := httputils.BoolValue(r, "stdout"), httputils.BoolValue(r, "stderr")
	if !(stdout || stderr) {
		return fmt.Errorf("Bad parameters: you must choose at least one stream")
	}

	var since time.Time
	if r.Form.Get("since") != "" {
		s, n, err := timetypes.ParseTimestamps(r.Form.Get("since"), 0)
		if err != nil {
			return err
		}
		since = time.Unix(s, n)
	}

	var closeNotifier <-chan bool
	if notifier, ok := w.(http.CloseNotifier); ok {
		closeNotifier = notifier.CloseNotify()
	}

	containerName := vars["name"]

	if !s.backend.Exists(containerName) {
		return derr.ErrorCodeNoSuchContainer.WithArgs(containerName)
	}

	// write an empty chunk of data (this is to ensure that the
	// HTTP Response is sent immediately, even if the container has
	// not yet produced any data)
	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	output := ioutils.NewWriteFlusher(w)
	defer output.Close()

	logsConfig := &daemon.ContainerLogsConfig{
		Follow:     httputils.BoolValue(r, "follow"),
		Timestamps: httputils.BoolValue(r, "timestamps"),
		Since:      since,
		Tail:       r.Form.Get("tail"),
		UseStdout:  stdout,
		UseStderr:  stderr,
		OutStream:  output,
		Stop:       closeNotifier,
	}

	if err := s.backend.ContainerLogs(containerName, logsConfig); err != nil {
		// The client may be expecting all of the data we're sending to
		// be multiplexed, so send it through OutStream, which will
		// have been set up to handle that if needed.
		fmt.Fprintf(logsConfig.OutStream, "Error running logs job: %s\n", utils.GetErrorMessage(err))
	}

	return nil
}

func (s *containerRouter) getContainersExport(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return s.backend.ContainerExport(vars["name"], w)
}

func (s *containerRouter) postContainersStart(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	// If contentLength is -1, we can assumed chunked encoding
	// or more technically that the length is unknown
	// https://golang.org/src/pkg/net/http/request.go#L139
	// net/http otherwise seems to swallow any headers related to chunked encoding
	// including r.TransferEncoding
	// allow a nil body for backwards compatibility
	var hostConfig *container.HostConfig
	if r.Body != nil && (r.ContentLength > 0 || r.ContentLength == -1) {
		if err := httputils.CheckForJSON(r); err != nil {
			return err
		}

		c, err := runconfig.DecodeHostConfig(r.Body)
		if err != nil {
			return err
		}

		hostConfig = c
	}

	if err := s.backend.ContainerStart(vars["name"], hostConfig); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *containerRouter) postContainersStop(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	seconds, _ := strconv.Atoi(r.Form.Get("t"))

	if err := s.backend.ContainerStop(vars["name"], seconds); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *containerRouter) postContainersKill(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	var sig syscall.Signal
	name := vars["name"]

	// If we have a signal, look at it. Otherwise, do nothing
	if sigStr := r.Form.Get("signal"); sigStr != "" {
		var err error
		if sig, err = signal.ParseSignal(sigStr); err != nil {
			return err
		}
	}

	if err := s.backend.ContainerKill(name, uint64(sig)); err != nil {
		theErr, isDerr := err.(errcode.ErrorCoder)
		isStopped := isDerr && theErr.ErrorCode() == derr.ErrorCodeNotRunning

		// Return error that's not caused because the container is stopped.
		// Return error if the container is not running and the api is >= 1.20
		// to keep backwards compatibility.
		version := httputils.VersionFromContext(ctx)
		if version.GreaterThanOrEqualTo("1.20") || !isStopped {
			return fmt.Errorf("Cannot kill container %s: %v", name, utils.GetErrorMessage(err))
		}
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *containerRouter) postContainersRestart(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	timeout, _ := strconv.Atoi(r.Form.Get("t"))

	if err := s.backend.ContainerRestart(vars["name"], timeout); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *containerRouter) postContainersPause(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := s.backend.ContainerPause(vars["name"]); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *containerRouter) postContainersUnpause(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := s.backend.ContainerUnpause(vars["name"]); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *containerRouter) postContainersWait(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	status, err := s.backend.ContainerWait(vars["name"], -1*time.Second)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, &types.ContainerWaitResponse{
		StatusCode: status,
	})
}

func (s *containerRouter) getContainersChanges(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	changes, err := s.backend.ContainerChanges(vars["name"])
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, changes)
}

func (s *containerRouter) getContainersTop(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	procList, err := s.backend.ContainerTop(vars["name"], r.Form.Get("ps_args"))
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, procList)
}

func (s *containerRouter) postContainerRename(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	name := vars["name"]
	newName := r.Form.Get("name")
	if err := s.backend.ContainerRename(name, newName); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *containerRouter) postContainerUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}
	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	var updateConfig container.UpdateConfig

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updateConfig); err != nil {
		return err
	}

	hostConfig := &container.HostConfig{
		Resources: updateConfig.Resources,
	}

	name := vars["name"]
	warnings, err := s.backend.ContainerUpdate(name, hostConfig)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusOK, &types.ContainerUpdateResponse{
		Warnings: warnings,
	})
}

func (s *containerRouter) postContainersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}
	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	name := r.Form.Get("name")

	config, hostConfig, networkingConfig, err := runconfig.DecodeContainerConfig(r.Body)
	if err != nil {
		return err
	}
	version := httputils.VersionFromContext(ctx)
	adjustCPUShares := version.LessThan("1.19")

	ccr, err := s.backend.ContainerCreate(types.ContainerCreateConfig{
		Name:             name,
		Config:           config,
		HostConfig:       hostConfig,
		NetworkingConfig: networkingConfig,
		AdjustCPUShares:  adjustCPUShares,
	})
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusCreated, ccr)
}

func (s *containerRouter) deleteContainers(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	name := vars["name"]
	config := &types.ContainerRmConfig{
		ForceRemove:  httputils.BoolValue(r, "force"),
		RemoveVolume: httputils.BoolValue(r, "v"),
		RemoveLink:   httputils.BoolValue(r, "link"),
	}

	if err := s.backend.ContainerRm(name, config); err != nil {
		// Force a 404 for the empty string
		if strings.Contains(strings.ToLower(err.Error()), "prefix can't be empty") {
			return fmt.Errorf("no such container: \"\"")
		}
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *containerRouter) postContainersResize(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	height, err := strconv.Atoi(r.Form.Get("h"))
	if err != nil {
		return err
	}
	width, err := strconv.Atoi(r.Form.Get("w"))
	if err != nil {
		return err
	}

	return s.backend.ContainerResize(vars["name"], height, width)
}

func (s *containerRouter) postContainersAttach(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	err := httputils.ParseForm(r)
	if err != nil {
		return err
	}
	containerName := vars["name"]

	_, upgrade := r.Header["Upgrade"]

	keys := []byte{}
	detachKeys := r.FormValue("detachKeys")
	if detachKeys != "" {
		keys, err = term.ToBytes(detachKeys)
		if err != nil {
			logrus.Warnf("Invalid escape keys provided (%s) using default : ctrl-p ctrl-q", detachKeys)
		}
	}

	attachWithLogsConfig := &daemon.ContainerAttachWithLogsConfig{
		Hijacker:   w.(http.Hijacker),
		Upgrade:    upgrade,
		UseStdin:   httputils.BoolValue(r, "stdin"),
		UseStdout:  httputils.BoolValue(r, "stdout"),
		UseStderr:  httputils.BoolValue(r, "stderr"),
		Logs:       httputils.BoolValue(r, "logs"),
		Stream:     httputils.BoolValue(r, "stream"),
		DetachKeys: keys,
	}

	return s.backend.ContainerAttachWithLogs(containerName, attachWithLogsConfig)
}

func (s *containerRouter) wsContainersAttach(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}
	containerName := vars["name"]

	if !s.backend.Exists(containerName) {
		return derr.ErrorCodeNoSuchContainer.WithArgs(containerName)
	}

	var keys []byte
	var err error
	detachKeys := r.FormValue("detachKeys")
	if detachKeys != "" {
		keys, err = term.ToBytes(detachKeys)
		if err != nil {
			logrus.Warnf("Invalid escape keys provided (%s) using default : ctrl-p ctrl-q", detachKeys)
		}
	}

	h := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		wsAttachWithLogsConfig := &daemon.ContainerWsAttachWithLogsConfig{
			InStream:   ws,
			OutStream:  ws,
			ErrStream:  ws,
			Logs:       httputils.BoolValue(r, "logs"),
			Stream:     httputils.BoolValue(r, "stream"),
			DetachKeys: keys,
		}

		if err := s.backend.ContainerWsAttachWithLogs(containerName, wsAttachWithLogsConfig); err != nil {
			logrus.Errorf("Error attaching websocket: %s", utils.GetErrorMessage(err))
		}
	})
	ws := websocket.Server{Handler: h, Handshake: nil}
	ws.ServeHTTP(w, r)

	return nil
}

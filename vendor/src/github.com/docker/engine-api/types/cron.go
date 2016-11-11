package types

import (
	"time"

	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/network"
)

type Cron struct {
	// Job name. Must be unique, acts as the id.
	Name string `json:"name"`

	// Cron expression for the job. When to run the job.
	Schedule string `json:"schedule"`

	ContainerName string                    `json:"container_name"`
	Config        *container.Config         `json:"config"`
	HostConfig    *container.HostConfig     `json:"host_config"`
	NetConfig     *network.NetworkingConfig `json:"net_config"`

	// Owner of the job.
	Owner string `json:"owner"`

	// Owner email of the job.
	OwnerEmail string `json:"owner_email"`

	// Number of successful executions of this job.
	SuccessCount int `json:"success_count"`

	// Number of errors running this job.
	ErrorCount int `json:"error_count"`

	// Last time this job executed succesful.
	LastSuccess time.Time `json:"last_success"`

	// Last time this job failed.
	LastError time.Time `json:"last_error"`

	// Is this job disabled?
	Disabled bool `json:"disabled"`

	// Tags of the target servers to run this job against.
	Tags map[string]string `json:"tags"`
}

type CronListOptions struct {
	Filters filters.Args
}

type Event struct {
	StartedAt  int64  `json:"StartedAt"`
	FinishedAt int64  `json:"FinishedAt"`
	Status     string `json:"Status"`
	Job        string `json:"Job"`
	Container  string `json:"Container"`
	Message    string `json:"Message"`
}

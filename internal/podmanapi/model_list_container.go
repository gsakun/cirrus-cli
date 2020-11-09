/*
 * Provides a container compatible interface.
 *
 * This documentation describes the Podman v2.0 RESTful API. It replaces the Podman v1.0 API and was initially delivered along with Podman v2.0.  It consists of a Docker-compatible API and a Libpod API providing support for Podman’s unique features such as pods.  To start the service and keep it running for 5,000 seconds (-t 0 runs forever):  podman system service -t 5000 &  You can then use cURL on the socket using requests documented below.  NOTE: if you install the package podman-docker, it will create a symbolic link for /var/run/docker.sock to /run/podman/podman.sock  See podman-service(1) for more information.  Quick Examples:  'podman info'  curl --unix-socket /run/podman/podman.sock http://d/v1.0.0/libpod/info  'podman pull quay.io/containers/podman'  curl -XPOST --unix-socket /run/podman/podman.sock -v 'http://d/v1.0.0/images/create?fromImage=quay.io%2Fcontainers%2Fpodman'  'podman list images'  curl --unix-socket /run/podman/podman.sock -v 'http://d/v1.0.0/libpod/images/json' | jq
 *
 * API version: 0.0.1
 * Contact: podman@lists.podman.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

// Listcontainer describes a container suitable for listing
type ListContainer struct {
	// Container command
	Command []string `json:"Command,omitempty"`
	// Container creation time
	Created int64 `json:"Created,omitempty"`
	// Human readable container creation time.
	CreatedAt string `json:"CreatedAt,omitempty"`
	// If container has exited, the return code from the command
	ExitCode int32 `json:"ExitCode,omitempty"`
	// If container has exited/stopped
	Exited bool `json:"Exited,omitempty"`
	// Time container exited
	ExitedAt int64 `json:"ExitedAt,omitempty"`
	// The unique identifier for the container
	Id string `json:"Id,omitempty"`
	// Container image
	Image string `json:"Image,omitempty"`
	// Container image ID
	ImageID string `json:"ImageID,omitempty"`
	// If this container is a Pod infra container
	IsInfra bool `json:"IsInfra,omitempty"`
	// Labels for container
	Labels map[string]string `json:"Labels,omitempty"`
	// User volume mounts
	Mounts []string `json:"Mounts,omitempty"`
	// The names assigned to the container
	Names []string `json:"Names,omitempty"`
	Namespaces *ListContainerNamespaces `json:"Namespaces,omitempty"`
	// The process id of the container
	Pid int64 `json:"Pid,omitempty"`
	// If the container is part of Pod, the Pod ID. Requires the pod boolean to be set
	Pod string `json:"Pod,omitempty"`
	// If the container is part of Pod, the Pod name. Requires the pod boolean to be set
	PodName string `json:"PodName,omitempty"`
	// Port mappings
	Ports []PortMapping `json:"Ports,omitempty"`
	Size *ContainerSize `json:"Size,omitempty"`
	// Time when container started
	StartedAt int64 `json:"StartedAt,omitempty"`
	// State of container
	State string `json:"State,omitempty"`
	// Status is a human-readable approximation of a duration for json output
	Status string `json:"Status,omitempty"`
}
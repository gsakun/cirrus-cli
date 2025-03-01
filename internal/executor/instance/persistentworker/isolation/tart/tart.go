package tart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cirruslabs/cirrus-cli/internal/executor/instance/persistentworker/remoteagent"
	"github.com/cirruslabs/cirrus-cli/internal/executor/instance/runconfig"
	"github.com/cirruslabs/cirrus-cli/internal/executor/platform"
	"github.com/cirruslabs/cirrus-cli/internal/logger"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	ErrFailed     = errors.New("tart isolation failed")
	ErrSyncFailed = errors.New("failed to sync project directory")
)

const (
	vmNamePrefix                = "cirrus-cli-"
	macOSAutomountDirectoryPath = "/Volumes/My Shared Files"
	macOSAutomountDirectoryItem = "working-dir"
)

type Tart struct {
	logger      logger.Lightweight
	vmName      string
	sshUser     string
	sshPassword string
	cpu         uint32
	memory      uint32
	softnet     bool
}

func New(vmName string, sshUser string, sshPassword string, cpu uint32, memory uint32, opts ...Option) (*Tart, error) {
	tart := &Tart{
		vmName:      vmName,
		sshUser:     sshUser,
		sshPassword: sshPassword,
		cpu:         cpu,
		memory:      memory,
	}

	// Apply options
	for _, opt := range opts {
		opt(tart)
	}

	// Apply default options (to cover those that weren't specified)
	if tart.logger == nil {
		tart.logger = &logger.LightweightStub{}
	}

	return tart, nil
}

func (tart *Tart) Run(ctx context.Context, config *runconfig.RunConfig) (err error) {
	tmpVMName := fmt.Sprintf("%s%d-", vmNamePrefix, config.TaskID) + uuid.NewString()
	vm, err := NewVMClonedFrom(ctx,
		tart.vmName, tmpVMName,
		tart.cpu, tart.memory,
		config.TartOptions.LazyPull,
		config.Logger(),
	)
	if err != nil {
		return fmt.Errorf("%w: failed to create VM cloned from %q: %v", ErrFailed, tart.vmName, err)
	}
	defer vm.Close()

	// Start the VM (asynchronously)
	var directoryMounts []directoryMount
	if config.DirtyMode {
		directoryMounts = append(directoryMounts, directoryMount{
			Name:     macOSAutomountDirectoryItem,
			Path:     config.ProjectDir,
			ReadOnly: false,
		})
	}

	if len(config.TartDirs) != 0 {
		for _, i := range config.TartDirs {
			field := strings.Split(i, ":")
			if len(field) < 2 {
				os.Exit(1)
			}
			if len(field) == 2 {
				directoryMounts = append(directoryMounts, directoryMount{
					Name: field[0],
					Path: field[1],
				})
			}
			if len(field) == 2 {
				if field[2] == "ro" {
					directoryMounts = append(directoryMounts, directoryMount{
						Name:     field[0],
						Path:     field[1],
						ReadOnly: true,
					})
					continue
				}
				if field[2] == "rw" {
					directoryMounts = append(directoryMounts, directoryMount{
						Name:     field[0],
						Path:     field[1],
						ReadOnly: false,
					})
					continue
				}
				fmt.Printf("Check the input read and write permissions for the mount directory")
				os.Exit(1)
			}
		}
	}

	vm.Start(tart.softnet, directoryMounts)

	// Wait for the VM to start and get it's DHCP address
	var ip string
	bootLogger := config.Logger().Scoped("boot virtual machine")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-vm.ErrChan():
			return err
		default:
			time.Sleep(time.Second)
		}

		ip, err = vm.RetrieveIP(ctx)
		if err != nil {
			tart.logger.Debugf("failed to retrieve VM %s IP: %v\n", vm.Ident(), err)
			continue
		}

		break
	}

	tart.logger.Debugf("IP %s retrieved from VM %s, running agent...", ip, vm.Ident())

	bootLogger.Errorf("VM was assigned with %s IP", ip)
	bootLogger.Finish(true)

	var hooks []remoteagent.WaitForAgentHook
	if config.ProjectDir != "" && !config.DirtyMode {
		hooks = append(hooks, func(ctx context.Context, sshClient *ssh.Client) error {
			syncLogger := config.Logger().Scoped("syncing working directory")
			if err := tart.syncProjectDir(config.ProjectDir, sshClient); err != nil {
				syncLogger.Finish(false)
				return fmt.Errorf("%w: %v", ErrSyncFailed, err)
			}

			syncLogger.Finish(true)
			return nil
		})
	}

	return remoteagent.WaitForAgent(ctx, tart.logger, ip,
		tart.sshUser, tart.sshPassword, "darwin", "arm64",
		config, true, hooks)
}

func (tart *Tart) WorkingDirectory(projectDir string, dirtyMode bool) string {
	if dirtyMode {
		return path.Join(macOSAutomountDirectoryPath, macOSAutomountDirectoryItem)
	}

	return platform.NewUnix().GenericWorkingDir()
}

func (tart *Tart) Close() error {
	return nil
}

func Cleanup() error {
	stdout, _, err := Cmd(context.Background(), "list", "--quiet")
	if err != nil {
		return err
	}

	for _, vmName := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if !strings.HasPrefix(vmName, vmNamePrefix) {
			continue
		}

		if _, _, err := Cmd(context.Background(), "delete", vmName); err != nil {
			return err
		}
	}

	return nil
}

func (tart *Tart) syncProjectDir(dir string, sshClient *ssh.Client) error {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		// Handle possible error that occurred when reading this directory entry information
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		remotePath := sftp.Join(platform.NewUnix().GenericWorkingDir(), relativePath)

		if fileInfo.Mode().IsDir() {
			return sftpClient.MkdirAll(remotePath)
		} else if fileInfo.Mode().IsRegular() {
			localFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer localFile.Close()

			remoteFile, err := sftpClient.Create(remotePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()

			if _, err := io.Copy(remoteFile, localFile); err != nil {
				return err
			}
		}

		return nil
	})
}

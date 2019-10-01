package main

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/pv-cleaner-operator/flag"
	"github.com/giantswarm/pv-cleaner-operator/pkg/project"
	"github.com/giantswarm/pv-cleaner-operator/server"
	"github.com/giantswarm/pv-cleaner-operator/service"
)

const (
	notAvailable = "n/a"
)

var (
	f *flag.Flag = flag.New()
)

func panicOnErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
}

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainWithError() error {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return microerror.Maskf(err, "micrologger.New")
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			serviceConfig := service.Config{
				Flag:   f,
				Logger: newLogger,
				Viper:  v,

				Description: project.Description(),
				GitCommit:   project.GitSHA(),
				ProjectName: project.Name(),
				Source:      project.Source(),
				Version:     project.Version(),
			}

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "service.New")))
			}
			go newService.Boot()
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			serverConfig := server.Config{
				Logger:  newLogger,
				Service: newService,
				Viper:   v,
			}

			newServer, err = server.New(serverConfig)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "server.New")))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        newLogger,
			ServerFactory: newServerFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.CobraCommand().Execute()

	return nil
}

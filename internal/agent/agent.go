/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"github.com/tschaefer/finchctl/internal/target"
)

type Agent interface {
	Deploy() error
	Teardown() error
	Register(string, *RegisterData) ([]byte, error)
	List(string) (*[]ListData, error)
	Deregister(string, string) error
	Config(string, string) ([]byte, error)
	Update() error
}

type agent struct {
	target target.Target
	config string
	format target.Format
	dryRun bool
}

func New(config, targetUrl string, format target.Format, dryRun bool) (Agent, error) {
	target, err := target.NewTarget(targetUrl, format, dryRun)
	if err != nil {
		return nil, err
	}

	return &agent{
		target: target,
		config: config,
		format: format,
		dryRun: dryRun,
	}, nil
}

func (a *agent) Teardown() error {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	if err := a.requirementsAgent(); err != nil {
		return convertError(err, &TeardownAgentError{})
	}

	if err := a.teardownAgent(); err != nil {
		return err
	}

	return nil
}

func (a *agent) Deploy() error {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	if err := a.requirementsAgent(); err != nil {
		return err
	}

	machine, err := a.machineInfo()
	if err != nil {
		return err
	}

	if err := a.deployAgent(machine); err != nil {
		return err
	}

	return nil
}

func (a *agent) Register(service string, data *RegisterData) ([]byte, error) {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	config, err := a.registerAgent(service, data)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (a *agent) List(service string) (*[]ListData, error) {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	return a.listAgents(service)
}

func (a *agent) Deregister(service, resourceID string) error {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	return a.deregisterAgent(service, resourceID)
}

func (a *agent) Config(service, resourceID string) ([]byte, error) {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	return a.configAgent(service, resourceID)
}

func (a *agent) Update() error {
	defer func() {
		if a.format == target.FormatProgress {
			println()
		}
	}()

	if err := a.requirementsAgent(); err != nil {
		return convertError(err, &UpdateAgentError{})
	}

	machine, err := a.machineInfo()
	if err != nil {
		return err
	}

	return a.updateAgent(machine)
}

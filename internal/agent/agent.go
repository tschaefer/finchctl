/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"github.com/tschaefer/finchctl/internal/target"
)

type Agent struct {
	target target.Target
	config string
	format target.Format
	dryRun bool
}

func New(config, targetUrl string, format target.Format, dryRun bool) (*Agent, error) {
	target, err := target.NewTarget(targetUrl, format, dryRun)
	if err != nil {
		return nil, err
	}

	return &Agent{
		target: target,
		config: config,
		format: format,
		dryRun: dryRun,
	}, nil
}

func (a *Agent) Teardown() error {
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

func (a *Agent) Deploy() error {
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

func (a *Agent) Register(service string, data *RegisterData) ([]byte, error) {
	return a.registerAgent(service, data)
}

func (a *Agent) List(service string) (*[]ListData, error) {
	return a.listAgents(service)
}

func (a *Agent) Deregister(service, resourceID string) error {
	return a.deregisterAgent(service, resourceID)
}

func (a *Agent) Config(service, resourceID string) ([]byte, error) {
	return a.configAgent(service, resourceID)
}

func (a *Agent) Update(skipConfig bool, skipBinaries bool) error {
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

	return a.updateAgent(machine, skipConfig, skipBinaries)
}

func (a *Agent) Describe(service, resourceID string) (*DescribeData, error) {
	return a.describeAgent(service, resourceID)
}

func (a *Agent) Edit(service string, data *EditData) error {
	return a.editAgent(service, data)
}

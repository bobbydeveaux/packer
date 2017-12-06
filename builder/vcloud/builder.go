// The vcloud package contains a packer.Builder implementation
// that builds VMWare vCloud images (vApp templates).

package vcloud

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// The unique id for the builder
const BuilderId = "rickard-von-essen.vcloud"

// Configuration tells the builder the credentials
// to use while communicating with vCloud and describes the image
// you are creating
type config struct {
	common.PackerConfig `mapstructure:",squash"`

	BaseURL        string `mapstructure:"base_url"`
	Organization   string `mapstructure:"organization"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	SourceCatalog  string `mapstructure:"source_catalog"`
	SourceTemplate string `mapstructure:"source_template"`
	Catalog        string `mapstructure:"catalog"`
	TemplateName   string `mapstructure:"template_name"`
	NATSSH         bool   `mapstructure:"nat_ssh"`

	SSHUsername string `mapstructure:"ssh_username"`
	SSHPassword string `mapstructure:"ssh_password"`
	SSHPort     uint   `mapstructure:"ssh_port"`

	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	RawStateTimeout string `mapstructure:"state_timeout"`

	// These are unexported since they're set by other fields
	// being set.
	sshTimeout   time.Duration
	stateTimeout time.Duration

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.TemplateName == "" {
		// Default to packer-{{ unix timestamp (utc) }}
		b.config.TemplateName = "packer-{{timestamp}}"
	}

	if b.config.SSHUsername == "" {
		// Default to "root". You can override this if your
		// SourceTemplate VM has a different user account
		b.config.SSHUsername = "root"
	}

	if b.config.SSHPort == 0 {
		// Default to port 22
		b.config.SSHPort = 22
	}

	if b.config.RawSSHTimeout == "" {
		// Default to 1 minute timeouts
		b.config.RawSSHTimeout = "1m"
	}

	if b.config.RawStateTimeout == "" {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for vApp to become active
		b.config.RawStateTimeout = "6m"
	}

	templates := map[string]*string{

		"source_catalog":  &b.config.SourceCatalog,
		"source_template": &b.config.SourceTemplate,
		"catalog":         &b.config.Catalog,
		"template_name":   &b.config.TemplateName,
		"ssh_username":    &b.config.SSHUsername,
		"ssh_timeout":     &b.config.RawSSHTimeout,
		"state_timeout":   &b.config.RawStateTimeout,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	// TODO: change this!
	if b.config.BaseURL == "" {
		b.config.BaseURL = "https://api.vmware.com" // TODO vCloud Air as default!
	}
	_, err = url.Parse(b.config.BaseURL)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing base_url \"%s\": %s", b.config.BaseURL, err))
	}

	sshTimeout, err := time.ParseDuration(b.config.RawSSHTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}
	b.config.sshTimeout = sshTimeout

	stateTimeout, err := time.ParseDuration(b.config.RawStateTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing state_timeout: %s", err))
	}
	b.config.stateTimeout = stateTimeout

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	common.ScrubConfig(b.config, b.config.Password)
	common.ScrubConfig(b.config, b.config.SSHPassword)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

	baseUrl, _ := url.Parse(b.config.BaseURL)
	client := NewClient(
		*baseUrl,
		b.config.Organization,
		b.config.Username,
		b.config.Password,
	)

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("client", &client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		new(stepLogin),
		new(stepCreateVApp),
		new(stepStart),
		new(stepNATSSH),
		//new(stepCreateSSHKey),
		//		new(stepCreateDroplet),
		//		new(stepDropletInfo),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: 5 * time.Minute,
		},
		new(common.StepProvision),
		//new(common.StepShutdown),
		new(stepStop),
		//		new(stepSnapshot),
	}

	// Run the steps
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("template_name"); !ok {
		log.Println("Failed to find template_name in state. Bug?")
		return nil, nil
	}

	//	sregion := state.Get("region")
	//
	//	var region string
	//
	//	if sregion != nil {
	//		region = sregion.(string)
	//	} else {
	//		region = fmt.Sprintf("%v", state.Get("region_id").(uint))
	//	}
	//
	//	found_region, err := client.Region(region)
	//
	//	if err != nil {
	//		return nil, err
	//	}

	artifact := &Artifact{
		templateName: state.Get("template_name").(string),
		templateUid:  state.Get("template_id").(uint),
		client:       client,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

package vcloud

import (
	"fmt"
	"github.com/vmware/govcloudair"
	"log"
	"net/url"
)

type VCloudClient struct {
	BaseURL      url.URL
	Organization string
	User         string
	Password     string
	client       *govcloudair.VCDClient
	org          *govcloudair.Org
	vdc          *govcloudair.Vdc
}

func NewClient(BaseURL url.URL, Organization, User, Password string) VCloudClient {
	return VCloudClient{
		BaseURL:      BaseURL,
		Organization: Organization,
		User:         User,
		Password:     Password,
		client:       govcloudair.NewVCDClient(BaseURL),
		org:          &govcloudair.Org{},
		vdc:          &govcloudair.Vdc{},
	}
}

func (c *VCloudClient) Login() error {
	log.Printf("Login to: %v with %s@%s..", c.BaseURL, c.User, c.Organization)
	org, vdc, err := c.client.Authenticate(c.User, c.Password, c.Organization)
	if err != nil {
		return err
	}
	log.Printf("Login successfully!")
	c.org = &org
	c.vdc = &vdc
	return nil
}

func (c *VCloudClient) Logout() error {
	log.Printf("Logout from: %v with %s@%s..", c.BaseURL, c.User, c.Organization)
	err := c.client.Disconnect()
	if err != nil {
		return err
	}
	log.Printf("Logout successfully!")
	c.org = &govcloudair.Org{}
	c.vdc = &govcloudair.Vdc{}
	return nil
}

func (c *VCloudClient) CreateVApp(templateName, sourceCatalog, sourceTemplate string) (string, error) {

	// create vApp from catalog/source_template
	log.Printf("Finding catalog %s.\n", sourceCatalog)
	cat, err := c.org.FindCatalog(sourceCatalog)
	if err != nil {
		return "", fmt.Errorf("Error trying to find catalog: \"%s\", %v\n", sourceCatalog, err)
	}

	log.Printf("Finding catalog item: %s.\n", sourceTemplate)
	ci, err := cat.FindCatalogItem(sourceTemplate)
	if err != nil {
		return "", fmt.Errorf("Error trying to find vApp template \"%s\" in catalog \"%s\", %v\n", sourceTemplate, sourceCatalog, err)
	}

	vappTmpl, err := ci.GetVAppTemplate()
	if err != nil {
		return "", fmt.Errorf("Error getting vApp Template from catalog item \"%s\", %v\n", sourceTemplate, err)
	}

	// TODO: network
	log.Printf("Finding vDC network: %s.\n", "Internal Network")
	net, err := c.vdc.FindVDCNetwork("Internal Network")
	if err != nil {
		return "", fmt.Errorf("Error finding vDC Network\"%s\", %v\n", "Internal Network", err)
	}

	vapp := govcloudair.NewVApp(&c.client.Client)
	// TODO: optional description.
	task, err := vapp.ComposeVApp(net, vappTmpl, templateName, "Packer built vApp Template.")

	err = task.Refresh()
	if err != nil {
		return "", fmt.Errorf("Error refresh task for vApp creation %v\n", err)
	}

	log.Printf("Waiting for vApp creation...\n")
	err = task.WaitTaskCompletion()
	if err != nil {
		return "", fmt.Errorf("Error waiting for vApp creation %v\n", err)
	}

	log.Printf("vApp creation completed\n")

	return vapp.VApp.ID, nil
}

func (c *VCloudClient) DestroyVApp(uid string) error {
	return nil
}

func (c *VCloudClient) StartVApp(uid string) error {
	vapp, err := c.vdc.FindVAppByID(uid)
	if err != nil {
		return fmt.Errorf("Error finding vApp %s, %v", uid, err)
	}

	task, err := vapp.PowerOn()
	if err != nil {
		return fmt.Errorf("Error powering on vApp %s, %v", vapp.VApp.Name, err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("Error wating for power on of vApp %s, %v", vapp.VApp.Name, err)
	}

	return nil
}

func (c *VCloudClient) PowerOffVApp(uid string) error {
	vapp, err := c.vdc.FindVAppByID(uid)
	if err != nil {
		return fmt.Errorf("Error finding vApp %s, %v", uid, err)
	}

	task, err := vapp.PowerOff()
	if err != nil {
		return fmt.Errorf("Error power off vApp %s, %v", vapp.VApp.Name, err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("Error wating for powering off vApp %s, %v", vapp.VApp.Name, err)
	}

	return nil
}

func (c *VCloudClient) GetVAppIP(uid string) (string, error) {
	vapp, err := c.vdc.FindVAppByID(uid)
	if err != nil {
		return "", fmt.Errorf("Error finding vApp %s, %v", uid, err)
	}

	ip := vapp.VApp.Children.VM[0].NetworkConnectionSection.NetworkConnection.IPAddress
	if ip == "" {
		return "", fmt.Errorf("No IP address found for VM in vApp %s", uid)
	}

	return ip, nil
}

func (c *VCloudClient) DNAT(vm_ip string, ssh_port uint) (string, uint, error) {
	// 81.91.13.X DNAT Gw ext IP:81.92.13.200 origin port 2223<free port> <vm ip> <ssh port> TCP (enabled)

	// FW src:external src_port:any dest:81.91.13.200 dest_port:2223 TCP allow
	gw, err := c.vdc.FindEdgeGateway("RickardvonEssen EGW")
	if err != nil {
		return "", 0, fmt.Errorf("Error finding Edge GW %s, %v", "RickardvonEssen EGW", err)
	}

	nat_rule := &types.NatRule{}
	nat_rule.RuleType = types.NatRule.RuleType.DNAT
	nat_rule.IsEnabled = true
	nat_rule.ID = "packer-vm-rule"
	nat_rule.GatewayNatRule = types.GatewayNatRule{}
	nat_rule.GatewayNatRule.Interface = "81.91.13.X"
	nat_rule.GatewayNatRule.OriginalIP = gw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.ExternalIP
	nat_rule.GatewayNatRule.OriginalPort = "2002" // free port
	nat_rule.GatewayNatRule.TranslatedIP = vm_ip
	nat_rule.GatewayNatRule.TranslatedPort = string(ssh_port)
	nat_rule.GatewayNatRule.Protocol = "Tcp"

	appendw.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule
}

func (c *VCloudClient) DestroyTemplate(uid uint) error {
	return nil
}

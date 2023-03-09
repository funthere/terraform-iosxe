package provider

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/meirizal/terraform-experiment/api/client"
	"github.com/meirizal/terraform-experiment/api/server"
)

func validateName(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("Expected name to be string"))
		return warns, errs
	}
	whiteSpace := regexp.MustCompile(`\s+`)
	if whiteSpace.Match([]byte(value)) {
		errs = append(errs, fmt.Errorf("host cannot contain whitespace. Got %s", value))
		return warns, errs
	}
	host_port := strings.Split(value, ":")
	if len(host_port) < 2 {
		errs = append(errs, fmt.Errorf("host must followed by :port. Got %s", value))
		return warns, errs
	}
	return warns, errs
}

func validateInterfaceType(v interface{}, k string) (ws []string, es []error) {

	var errs []error
	var warns []string
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("Expected name to be string"))
		return warns, errs
	}
	list := []string{"GigabitEthernet", "TwoGigabitEthernet", "FiveGigabitEthernet", "TenGigabitEthernet", "TwentyFiveGigE", "FortyGigabitEthernet", "HundredGigE", "TwoHundredGigE", "FourHundredGigE"}
	if !slices.Contains(list, value) {
		errs = append(errs, fmt.Errorf("Interface type is not valid. Got %s", value))
		return warns, errs
	}
	return warns, errs
}

func resourceItem() *schema.Resource {
	fmt.Print()
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The host to push config",
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of an item",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default is 'admin'",
				Default:     "admin",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default is 'admin'",
				Default:     "admin",
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "Interface type",
				Required:     true,
				ValidateFunc: validateInterfaceType,
			},
			"number": {
				Type:        schema.TypeString,
				Description: "Interface number",
				Default:     "0",
				Optional:    true,
			},
			"ipv4_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IPv4",
				RequiredWith: []string{"ipv4_address_mask"},
				ValidateFunc: validation.IsIPv4Address,
			},
			"ipv4_address_mask": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "IPv4 mask",
				ValidateFunc: validation.IsIPv4Address,
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum Transmission Unit",
			},
			"shutdown": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"service_policy_input": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Service Policy Input",
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			"service_policy_output": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Service Policy output",
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
		},
		Create: resourceCreateItem,
		Read:   resourceReadItem,
		Update: resourceUpdateItem,
		Delete: resourceDeleteItem,
		Exists: resourceExistsItem,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCreateItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	item := getItemData(d)

	err := apiClient.NewItem(&item)

	if err != nil {
		return err
	}
	d.SetId(item.Host)

	return nil
}

func resourceReadItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	itemId := d.Id()
	item, err := apiClient.GetItem(itemId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
		} else {
			return fmt.Errorf("error finding Item with ID %s", itemId)
		}
	}

	d.SetId(item.Host)
	d.Set("host", item.Host)
	d.Set("description", item.Description)
	d.Set("type", item.IntfType)
	d.Set("number", item.Number)
	d.Set("ipv4_address", item.Ipv4Address)
	d.Set("ipv4_address_mask", item.Ipv4AddressMask)
	d.Set("mtu", item.Mtu)
	d.Set("shutdown", item.Shutdown)
	d.Set("service_policy_input", item.ServicePolicyInput)
	d.Set("service_policy_output", item.ServicePolicyOutput)
	return nil
}

func resourceUpdateItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	item := getItemData(d)

	err := apiClient.UpdateItem(&item)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteItem(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	item := getItemData(d)

	err := apiClient.DeleteItem(&item)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceExistsItem(d *schema.ResourceData, m interface{}) (bool, error) {
	apiClient := m.(*client.Client)

	itemId := d.Id()
	_, err := apiClient.GetItem(itemId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func getItemData(d *schema.ResourceData) server.Item {
	item := server.Item{
		Host:                d.Get("host").(string),
		Description:         d.Get("description").(string),
		Username:            d.Get("username").(string),
		Password:            d.Get("password").(string),
		IntfType:            d.Get("type").(string),
		Number:              d.Get("number").(string),
		Ipv4Address:         d.Get("ipv4_address").(string),
		Ipv4AddressMask:     d.Get("ipv4_address_mask").(string),
		Mtu:                 d.Get("mtu").(int),
		Shutdown:            d.Get("shutdown").(bool),
		ServicePolicyInput:  d.Get("service_policy_input").(string),
		ServicePolicyOutput: d.Get("service_policy_output").(string),
	}

	return item
}

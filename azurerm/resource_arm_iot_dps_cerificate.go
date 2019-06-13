package azurerm

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/Azure/azure-sdk-for-go/services/provisioningservices/mgmt/2018-01-22/iothub"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmIotDPSCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmIotDPSCertificateCreateOrUpdate,
		Read:   resourceArmIotDPSCertificateRead,
		Update: resourceArmIotDPSCertificateCreateOrUpdate,
		Delete: resourceArmIotDPSCertificateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.IoTHubName,
			},

			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"iot_dps_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.IoTHubName,
			},

			"certificate_content": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 10000),
			},
		},
	}
}

func resourceArmIotDPSCertificateCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).iothub.DPSCertificateClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	iotDPSName := d.Get("iot_dps_name").(string)

	if requireResourcesToBeImported && d.IsNewResource() {
		existing, err := client.Get(ctx, name, resourceGroup, iotDPSName, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_iot_dps_certificate", *existing.ID)
		}
	}

	certificate := iothub.CertificateBodyDescription{
		Certificate: utils.String(d.Get("certificate_content").(string)),
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, iotDPSName, name, certificate, ""); err != nil {
		return fmt.Errorf("Error creating/updating IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}

	resp, err := client.Get(ctx, name, resourceGroup, iotDPSName, "")
	if err != nil {
		return fmt.Errorf("Error retrieving IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}

	if resp.ID == nil {
		return fmt.Errorf("Cannot read IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}

	d.SetId(*resp.ID)

	return resourceArmIotDPSCertificateRead(d, meta)
}

func resourceArmIotDPSCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).iothub.DPSCertificateClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	iotDPSName := id.Path["provisioningServices"]
	name := id.Path["certificates"]

	resp, err := client.Get(ctx, name, resourceGroup, iotDPSName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("iot_dps_name", iotDPSName)

	return nil
}

func resourceArmIotDPSCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).iothub.DPSCertificateClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	iotDPSName := id.Path["provisioningServices"]
	name := id.Path["certificates"]

	resp, err := client.Get(ctx, name, resourceGroup, iotDPSName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil
		}
		return fmt.Errorf("Error retrieving IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}

	// TODO address this delete call if https://github.com/Azure/azure-rest-api-specs/pull/6311 get's merged
	if _, err := client.Delete(ctx, resourceGroup, *resp.Etag, iotDPSName, name, "", nil, nil, iothub.ServerAuthentication, nil, nil, nil, ""); err != nil {
		return fmt.Errorf("Error deleting IoT Device Provisioning Service Certificate %q (Device Provisioning Service %q / Resource Group %q): %+v", name, iotDPSName, resourceGroup, err)
	}
	return nil
}

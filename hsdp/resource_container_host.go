package hsdp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/loafoe/easyssh-proxy/v2"
	"github.com/philips-software/go-hsdp-api/cartel"
	"os"

	"log"
	"net/http"
	"strings"
	"time"
)

const (
	fileField     = "file"
	commandsField = "commands"
)

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Required:         true,
		ValidateDiagFunc: validateTags,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// TODO: handle empty tags
			return k == "tags.billing"
		},
		DefaultFunc: func() (interface{}, error) {
			return map[string]interface{}{"billing": ""}, nil
		},
		Elem: &schema.Schema{Type: schema.TypeString},
	}
}

func validateTags(v interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	tagsMap, ok := v.(map[string]interface{})
	if !ok {
		return diag.FromErr(fmt.Errorf("expected %q to be a map", v))
	}
	if len(tagsMap) > 8 {
		return diag.FromErr(fmt.Errorf("maximum of 8 tags are supported"))
	}
	for k, v := range tagsMap {
		if strings.EqualFold(k, "name") {
			return diag.FromErr(fmt.Errorf("tag \"%s\" is reserved by the Cartel API", k))
		}
		val, ok := v.(string)
		if !ok {
			return diag.FromErr(fmt.Errorf("tag \"%s\" value is of type %q", k, v))
		}
		if len(val) > 255 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Detail:   fmt.Sprintf("value of tag \"%s\" is too long (max=255)", k),
			})
		}
	}
	return diags
}

func resourceContainerHost() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: resourceContainerHostCreate,
		ReadContext:   resourceContainerHostRead,
		UpdateContext: resourceContainerHostUpdate,
		DeleteContext: resourceContainerHostDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_role": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "container-host",
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "m5.large",
			},
			"volume_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"iops"},
			},
			"iops": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4000),
			},
			"protect": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"encrypt_volumes": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
				ForceNew: true,
			},
			"volumes": {
				Type:         schema.TypeInt,
				Default:      0,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 6),
			},
			"volume_size": {
				Type:         schema.TypeInt,
				Default:      0,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 1000),
			},
			"security_groups": {
				Type:     schema.TypeSet,
				MaxItems: 5,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_groups": {
				Type:     schema.TypeSet,
				MaxItems: 50,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"bastion_host": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"private_key"},
			},
			"private_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{"user"},
			},
			commandsField: {
				Type:     schema.TypeList,
				MaxItems: 10,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			fileField: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"content": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"destination": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"subnet_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"subnet"},
			},
			"subnet": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_devices": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
		SchemaVersion: 3,
	}
}

func InstanceStateRefreshFunc(client *cartel.Client, nameTag string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		state, resp, err := client.GetDeploymentState(nameTag)
		if err != nil {
			log.Printf("Error on InstanceStateRefresh: %s", err)
			return resp, "", err
		}

		for _, failState := range failStates {
			if state == failState {
				return resp, state, fmt.Errorf("failed to reach target state, reason: %s",
					state)
			}
		}
		return resp, state, nil
	}
}

func resourceContainerHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	client, err := config.CartelClient()
	if err != nil {
		return diag.FromErr(err)
	}

	tagName := d.Get("name").(string)
	protect := d.Get("protect").(bool)
	iops := d.Get("iops").(int)
	encryptVolumes := d.Get("encrypt_volumes").(bool)
	volumeSize := d.Get("volume_size").(int)
	numberOfVolumes := d.Get("volumes").(int)
	volumeType := d.Get("volume_type").(string)
	instanceType := d.Get("instance_type").(string)
	securityGroups := expandStringList(d.Get("security_groups").(*schema.Set).List())
	userGroups := expandStringList(d.Get("user_groups").(*schema.Set).List())
	instanceRole := d.Get("instance_role").(string)
	subnetType := d.Get("subnet_type").(string)
	bastionHost := d.Get("bastion_host").(string)
	if bastionHost == "" {
		bastionHost = client.BastionHost()
	}
	user := d.Get("user").(string)
	privateKey := d.Get("private_key").(string)

	if subnetType == "" {
		subnetType = "private"
	}
	subnet := d.Get("subnet").(string)
	tagList := d.Get("tags").(map[string]interface{})
	tags := make(map[string]string)
	for t, v := range tagList {
		if val, ok := v.(string); ok {
			tags[t] = val
		}
	}
	// Fetch files first before starting provisioning
	createFiles, diags := collectFilesToCreate(d)
	if len(diags) > 0 {
		return diags
	}
	// And commands
	commands, diags := collectCommands(d)
	if len(diags) > 0 {
		return diags
	}
	if len(commands) > 0 {
		if user == "" {
			return diag.FromErr(fmt.Errorf("user must be set when '%s' is specified", commandsField))
		}
		if privateKey == "" {
			return diag.FromErr(fmt.Errorf("privateKey must be set when '%s' is specified", commandsField))
		}
	}

	ch, resp, err := client.Create(tagName,
		cartel.SecurityGroups(securityGroups...),
		cartel.UserGroups(userGroups...),
		cartel.VolumeType(volumeType),
		cartel.IOPs(iops),
		cartel.InstanceType(instanceType),
		cartel.VolumesAndSize(numberOfVolumes, volumeSize),
		cartel.VolumeEncryption(encryptVolumes),
		cartel.Protect(protect),
		cartel.InstanceRole(instanceRole),
		cartel.SubnetType(subnetType),
		cartel.Tags(tags),
		cartel.InSubnet(subnet),
	)
	instanceID := ""
	ipAddress := ""
	if err != nil {
		if resp == nil {
			_, _, _ = client.Destroy(tagName)
			return diag.FromErr(fmt.Errorf("create error (resp=nil): %w", err))
		}
		if ch == nil || resp.StatusCode >= 500 { // Possible 504, or other timeout, try to recover!
			if details := findInstanceByName(client, tagName); details != nil {
				instanceID = details.InstanceID
				ipAddress = details.PrivateAddress
			} else {
				_, _, _ = client.Destroy(tagName)
				return diag.FromErr(fmt.Errorf("create error (status=%d): %w", resp.StatusCode, err))
			}
		} else {
			_, _, _ = client.Destroy(tagName)
			return diag.FromErr(fmt.Errorf("create error (description=[%s], code=[%d]): %w", ch.Description, resp.StatusCode, err))
		}
	} else {
		instanceID = ch.InstanceID()
		ipAddress = ch.IPAddress()
	}
	d.SetId(instanceID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"provisioning", "indeterminate"},
		Target:     []string{"succeeded"},
		Refresh:    InstanceStateRefreshFunc(client, tagName, []string{"failed", "terminated", "shutting-down"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		// Trigger a delete to prevent failed instances from lingering
		_, _, _ = client.Destroy(tagName)
		return diag.FromErr(fmt.Errorf(
			"error waiting for instance '%s' to become ready: %s",
			instanceID, err))
	}
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": ipAddress,
	})
	// Collect SSH details
	privateIP := ipAddress
	ssh := &easyssh.MakeConfig{
		User:   user,
		Server: privateIP,
		Port:   "22",
		Key:    privateKey,
		Proxy:  http.ProxyFromEnvironment,
		Bastion: easyssh.DefaultConfig{
			User:   user,
			Server: bastionHost,
			Port:   "22",
			Key:    privateKey,
		},
	}

	// Create files
	if err := copyFiles(ssh, config, createFiles); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "failed to copy all files",
			Detail:   fmt.Sprintf("One or more files failed to copy: %v", err),
		})
	}

	// Run commands
	for i := 0; i < len(commands); i++ {
		stdout, stderr, done, err := ssh.Run(commands[i], 5*time.Minute)
		if err != nil {
			return append(diags, diag.FromErr(fmt.Errorf("command [%s]: %w", commands[i], err))...)
		} else {
			_, _ = config.Debug("command: %s\ndone: %t\nstdout:\n%s\nstderr:\n%s\n", commands[i], done, stdout, stderr)
		}
	}
	readDiags := resourceContainerHostRead(ctx, d, m)
	return append(diags, readDiags...)
}

func findInstanceByName(client *cartel.Client, name string) *cartel.InstanceDetails {
	instances, _, err := client.GetAllInstances()
	if err != nil {
		return nil
	}
	for _, i := range *instances {
		if i.NameTag == name {
			return &i
		}
	}
	return nil
}

func copyFiles(ssh *easyssh.MakeConfig, config *Config, createFiles []provisionFile) error {
	for _, f := range createFiles {
		if f.Source != "" {
			src, srcErr := os.Open(f.Source)
			if srcErr != nil {
				_, _ = config.Debug("Failed to open source file %s: %v\n", f.Source, srcErr)
				return srcErr
			}
			srcStat, statErr := src.Stat()
			if statErr != nil {
				_, _ = config.Debug("Failed to stat source file %s: %v\n", f.Source, statErr)
				_ = src.Close()
				return statErr
			}
			_ = ssh.WriteFile(src, srcStat.Size(), f.Destination)
			_, _ = config.Debug("Copied %s to remote file %s:%s: %d bytes\n", f.Source, ssh.Server, f.Destination, srcStat.Size())
			_ = src.Close()
		} else {
			buffer := bytes.NewBufferString(f.Content)
			// Should we fail the complete provision on errors here?
			_ = ssh.WriteFile(buffer, int64(buffer.Len()), f.Destination)
			_, _ = config.Debug("Created remote file %s:%s: %d bytes\n", ssh.Server, f.Destination, len(f.Content))
		}
	}
	return nil
}

func collectCommands(d *schema.ResourceData) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	list := d.Get(commandsField).([]interface{})
	commands := make([]string, 0)
	for i := 0; i < len(list); i++ {
		commands = append(commands, list[i].(string))
	}
	return commands, diags

}

type provisionFile struct {
	Source      string
	Content     string
	Destination string
}

func collectFilesToCreate(d *schema.ResourceData) ([]provisionFile, diag.Diagnostics) {
	var diags diag.Diagnostics
	files := make([]provisionFile, 0)
	if v, ok := d.GetOk(fileField); ok {
		vL := v.(*schema.Set).List()
		for _, vi := range vL {
			mVi := vi.(map[string]interface{})
			file := provisionFile{
				Source:      mVi["source"].(string),
				Content:     mVi["content"].(string),
				Destination: mVi["destination"].(string),
			}
			if file.Source == "" && file.Content == "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "conflict in file block",
					Detail:   fmt.Sprintf("file %s has neither 'source' or 'content', set one", file.Destination),
				})
				continue
			}
			if file.Source != "" && file.Content != "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "conflict in file block",
					Detail:   fmt.Sprintf("file %s has conflicting 'source' and 'content', choose only one", file.Destination),
				})
				continue
			}
			if file.Source != "" {
				src, srcErr := os.Open(file.Source)
				if srcErr != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "issue with source",
						Detail:   fmt.Sprintf("file %s: %v", file.Source, srcErr),
					})
					continue
				}
				_, statErr := src.Stat()
				if statErr != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "issue with source stat",
						Detail:   fmt.Sprintf("file %s: %v", file.Source, statErr),
					})
					_ = src.Close()
					continue
				}
				_ = src.Close()
			}
			files = append(files, file)
		}
	}
	return files, diags
}

func resourceContainerHostUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.CartelClient()
	if err != nil {
		return diag.FromErr(err)
	}

	tagName := d.Get("name").(string)
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return diag.FromErr(err)
	}
	if ch.InstanceID != d.Id() {
		return diag.FromErr(ErrInstanceIDMismatch)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		change := generateTagChange(o, n)
		log.Printf("[o:%v] [n:%v] [c:%v]\n", o, n, change)
		_, _, err := client.AddTags([]string{tagName}, change)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("user_groups") {
		o, n := d.GetChange("user_groups")
		old := expandStringList(o.(*schema.Set).List())
		newEntries := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newEntries, old)
		toRemove := difference(old, newEntries)

		// Additions
		if len(toAdd) > 0 {
			_, _, err := client.AddUserGroups([]string{tagName}, toAdd)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Removals
		if len(toRemove) > 0 {
			_, _, err := client.RemoveUserGroups([]string{tagName}, toRemove)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("security_groups") {
		o, n := d.GetChange("security_groups")
		old := expandStringList(o.(*schema.Set).List())
		newEntries := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newEntries, old)
		toRemove := difference(old, newEntries)

		// Additions
		if len(toAdd) > 0 {
			_, _, err := client.AddSecurityGroups([]string{tagName}, toAdd)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Removals
		if len(toRemove) > 0 {
			_, _, err := client.RemoveSecurityGroups([]string{tagName}, toRemove)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if d.HasChange("protect") {
		protect := d.Get("protect").(bool)
		_, _, err := client.SetProtection(tagName, protect)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return diags

}

func resourceContainerHostRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.CartelClient()
	if err != nil {
		return diag.FromErr(err)
	}

	tagName := d.Get("name").(string)

	if tagName == "" { // This is an import, find and set the tagName
		instances, _, err := client.GetAllInstances()
		if err != nil {
			return diag.FromErr(fmt.Errorf("cartel.GetAllInstances: %w", err))
		}
		id := d.Id()
		_ = d.Set("encrypt_volumes", true)
		_ = d.Set("volume_size", 0)
		for _, i := range *instances {
			if i.InstanceID == id {
				_ = d.Set("name", i.NameTag)
				tagName = i.NameTag
				break
			}
		}
	}

	state, resp, err := client.GetDeploymentState(tagName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			// State not found, probably a botched provision :(
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if state != "succeeded" {
		// Unless we have a succeeded deploy, taint the resource
		d.SetId("")
		return diags
	}
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return diag.FromErr(err)
	}
	if ch.InstanceID != d.Id() {
		return diag.FromErr(ErrInstanceIDMismatch)
	}
	_ = d.Set("protect", ch.Protection)
	_ = d.Set("volumes", len(ch.BlockDevices)-1) // -1 for the root volume
	_ = d.Set("role", ch.Role)
	_ = d.Set("launch_time", ch.LaunchTime)
	_ = d.Set("block_devices", ch.BlockDevices)
	_ = d.Set("security_groups", difference(ch.SecurityGroups, []string{"base"})) // Remove "base"
	_ = d.Set("user_groups", ch.LdapGroups)
	_ = d.Set("instance_type", ch.InstanceType)
	_ = d.Set("instance_role", ch.Role)
	_ = d.Set("vpc", ch.Vpc)
	_ = d.Set("zone", ch.Zone)
	_ = d.Set("launch_time", ch.LaunchTime)
	_ = d.Set("private_ip", ch.PrivateAddress)
	_ = d.Set("public_ip", ch.PublicAddress)
	_ = d.Set("subnet", ch.Subnet)
	subnetType := "private"
	if ch.PublicAddress != "" {
		subnetType = "public"
	}
	_ = d.Set("subnet_type", subnetType)
	_ = d.Set("tags", normalizeTags(ch.Tags))

	return diags
}

func resourceContainerHostDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)

	var diags diag.Diagnostics

	client, err := config.CartelClient()
	if err != nil {
		return diag.FromErr(err)
	}

	tagName := d.Get("name").(string)
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return diag.FromErr(err)
	}
	if ch.InstanceID != d.Id() {
		return diag.FromErr(ErrInstanceIDMismatch)
	}
	_, _, err = client.Destroy(tagName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags

}

func normalizeTags(tags map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range tags {
		if k == "billing" || v == "" {
			continue
		}
		normalized[k] = v
	}
	return normalized
}

func generateTagChange(old, new interface{}) map[string]string {
	change := make(map[string]string)
	o := old.(map[string]interface{})
	n := new.(map[string]interface{})
	for k := range o {
		if newVal, ok := n[k]; !ok || newVal == "" {
			change[k] = ""
		}
	}
	for k, v := range n {
		if k == "billing" {
			continue
		}
		if s, ok := v.(string); ok {
			change[k] = s
		}
	}
	return change
}

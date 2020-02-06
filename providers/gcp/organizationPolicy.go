package gcp

import (
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraform_utils"

	"golang.org/x/net/context"
	"google.golang.org/api/cloudresourcemanager/v1"
	cloudresourcemanagerv2 "google.golang.org/api/cloudresourcemanager/v2"
)

type OrganizationPolicyGenerator struct {
	GCPService
	OrganizationID string
}

// Run on routersList and create for each TerraformResource
func (g OrganizationPolicyGenerator) createResources(ctx context.Context, policiesList *cloudresourcemanager.OrganizationsListOrgPoliciesCall) []terraform_utils.Resource {
	resources := []terraform_utils.Resource{}
	if err := policiesList.Pages(ctx, func(page *cloudresourcemanager.ListOrgPoliciesResponse) error {
		for _, obj := range page.Policies {
			resources = append(resources, terraform_utils.NewSimpleResource(
				fmt.Sprintf("%s/%s", g.OrganizationID, obj.Constraint),
				fmt.Sprintf("%s-%s", g.OrganizationID, obj.Constraint),
				"google_organization_policy",
				"google",
				[]string{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

// Generate TerraformResources from GCP API,
// from each routers create 1 TerraformResource
// Need routers name as ID for terraform resource
func (g *OrganizationPolicyGenerator) InitResources() error {
	ctx := context.Background()
	service, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	serviceV2, err := cloudresourcemanagerv2.NewService(ctx)
	resource := "organizations/" + g.GetArgs()["project"].(string)
	resp, err := service.Projects.Get(g.GetArgs()["project"].(string)).Context(ctx).Do()
	if err != nil {
		return err
	}

	if resp.Parent == nil {
		return fmt.Errorf("project don't have parent")
	}
	folder, err := serviceV2.Folders.Get(resp.Parent.Id).Context(ctx).Do()
	if err != nil {
		return err
	}
	_, err = service.Organizations.Get(folder.Parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("don't find parent organization")
	}
	g.OrganizationID = folder.Parent
	policiesList := service.Organizations.ListOrgPolicies(resource, &cloudresourcemanager.ListOrgPoliciesRequest{
		PageSize: 100,
	})
	g.Resources = g.createResources(ctx, policiesList)
	return nil
}

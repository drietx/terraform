package terraform

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/configschema"
	"github.com/hashicorp/terraform/dag"
)

// ConcreteProviderNodeFunc is a callback type used to convert an
// abstract provider to a concrete one of some type.
type ConcreteProviderNodeFunc func(*NodeAbstractProvider) dag.Vertex

// NodeAbstractProvider represents a provider that has no associated operations.
// It registers all the common interfaces across operations for providers.
type NodeAbstractProvider struct {
	NameValue string
	PathValue []string

	// The fields below will be automatically set using the Attach
	// interfaces if you're running those transforms, but also be explicitly
	// set if you already have that information.

	Config *config.ProviderConfig
	Schema *configschema.Block
}

func (n *NodeAbstractProvider) Name() string {
	result := fmt.Sprintf("provider.%s", n.NameValue)
	if len(n.PathValue) > 1 {
		result = fmt.Sprintf("%s.%s", modulePrefixStr(n.PathValue), result)
	}

	return result
}

// GraphNodeSubPath
func (n *NodeAbstractProvider) Path() []string {
	return n.PathValue
}

// RemovableIfNotTargeted
func (n *NodeAbstractProvider) RemoveIfNotTargeted() bool {
	// We need to add this so that this node will be removed if
	// it isn't targeted or a dependency of a target.
	return true
}

// GraphNodeAttachSchema
func (n *NodeAbstractProvider) AttachSchema(schemas *Schemas) {
	providerName := providerSchemaKey(n.NameValue)
	providerSchemas := schemas.Providers[providerName]

	if providerSchemas == nil {
		log.Printf("[WARNING] No schemas available for provider %q", providerName)
		return
	}

	n.Schema = providerSchemas.Provider
	if n.Schema == nil {
		log.Printf("[WARNING] %s has no schema available", n.Name())
	}
}

// GraphNodeReferencer
func (n *NodeAbstractProvider) References() []string {
	if n.Config == nil {
		return nil
	}

	return ReferencesFromConfig(n.Config.RawConfig, n.Schema)
}

// GraphNodeProvider
func (n *NodeAbstractProvider) ProviderName() string {
	return n.NameValue
}

// GraphNodeProvider
func (n *NodeAbstractProvider) ProviderConfig() *config.RawConfig {
	if n.Config == nil {
		return nil
	}

	return n.Config.RawConfig
}

// GraphNodeAttachProvider
func (n *NodeAbstractProvider) AttachProvider(c *config.ProviderConfig) {
	n.Config = c
}

// GraphNodeDotter impl.
func (n *NodeAbstractProvider) DotNode(name string, opts *dag.DotOpts) *dag.DotNode {
	return &dag.DotNode{
		Name: name,
		Attrs: map[string]string{
			"label": n.Name(),
			"shape": "diamond",
		},
	}
}

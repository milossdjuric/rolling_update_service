package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/milossdjuric/rolling_update_service/internal/utils"
)

var (
	SupportedResourceQuotas = []string{"mem", "cpu", "disk"}
)

type App struct {
	Name           string
	SelectorLabels map[string]string
}

type AppSpec struct {
	Name                     string
	Namespace                string
	OrgId                    string
	SelectorLabels           map[string]string
	Quotas                   map[string]float64
	SeccompProfile           SeccompProfile
	SeccompDefintionStrategy string
}

type SeccompProfile struct {
	Version       string
	DefaultAction string
	Syscalls      []SyscallRule
}

type SyscallRule struct {
	Names  []string
	Action string
}

func NewAppSpec(name string, namespace string, orgId string, selectorLabels map[string]string, seccompProfile SeccompProfile, seccompDefintionStrategy string) AppSpec {

	profile := SeccompProfile{
		Version:       seccompProfile.Version,
		DefaultAction: seccompProfile.DefaultAction,
		Syscalls:      []SyscallRule{},
	}
	for _, syscall := range seccompProfile.Syscalls {
		rule := SyscallRule{
			Names:  syscall.Names,
			Action: syscall.Action,
		}
		profile.Syscalls = append(profile.Syscalls, rule)
	}

	appSpec := AppSpec{
		Name:                     name,
		Namespace:                namespace,
		OrgId:                    orgId,
		SeccompDefintionStrategy: seccompDefintionStrategy,
		SeccompProfile:           profile,
		Quotas:                   make(map[string]float64),
		SelectorLabels:           make(map[string]string),
	}
	for k, v := range selectorLabels {
		appSpec.SelectorLabels[k] = v
	}

	return appSpec
}

func (a *AppSpec) AddResourceQuota(resource string, quota float64) error {
	if !slices.Contains(SupportedResourceQuotas, resource) {
		return fmt.Errorf("quotas for a resource with name %s are not supported", resource)
	}
	a.Quotas[resource] = quota
	return nil
}

func (a *AppSpec) CompareAppSpecs(other AppSpec) bool {
	if a.Name != other.Name ||
		a.Namespace != other.Namespace ||
		a.OrgId != other.OrgId ||
		a.SeccompDefintionStrategy != other.SeccompDefintionStrategy ||
		a.SeccompProfile.DefaultAction != other.SeccompProfile.DefaultAction {
		return false
	}

	if !utils.MatchLabels(other.SelectorLabels, a.SelectorLabels) {
		return false
	}

	if !utils.CompareFloatMaps(a.Quotas, other.Quotas) {
		return false
	}

	if !a.CompareSyscalls(other.SeccompProfile.Syscalls) {
		return false
	}
	return true
}

func (a *AppSpec) CompareSyscalls(otherSyscalls []SyscallRule) bool {
	if len(a.SeccompProfile.Syscalls) != len(otherSyscalls) {
		return false
	}

	for i := range a.SeccompProfile.Syscalls {
		if a.SeccompProfile.Syscalls[i].Action != otherSyscalls[i].Action ||
			!utils.CompareStringSlices(a.SeccompProfile.Syscalls[i].Names, otherSyscalls[i].Names) {
			return false
		}
	}
	return true
}

func (as AppSpec) Validate() error {
	if as.Name == "" {
		return errors.New("app name is empty")
	}
	if as.Namespace == "" {
		return errors.New("app namespace is empty")
	}
	if as.OrgId == "" {
		return errors.New("app orgId is empty")
	}
	if len(as.SelectorLabels) == 0 {
		return errors.New("app selector labels are missing")
	}

	return nil
}

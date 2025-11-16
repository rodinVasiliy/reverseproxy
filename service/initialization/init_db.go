package initialization

import (
	"fmt"
	service "reverseproxy/service"
)

func InItDB(ps *service.PolicyService) error {
	defaultPolicy, err := GetDefaultPolicy(ps)
	if err != nil {
		return fmt.Errorf("failed to get default policy %s", err)
	}
	ps.LoadtPolicyToDB(defaultPolicy)

	return nil
}

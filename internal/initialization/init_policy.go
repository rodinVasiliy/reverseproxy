package initialization

import (
	policy "reverseproxy/internal/domain/policy"
)

func getDefaultPolicy() (*policy.Policy, error) {

	wl := "95.67.162.0/24" // для тестов
	defaultPolicy := policy.Policy{WL: []string{wl}, Name: DEFAULT_POLICY_NAME}
	return &defaultPolicy, nil
}

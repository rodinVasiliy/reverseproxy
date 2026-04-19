package policy

type PolicyInUseError struct {
	Webapps []string
}

func (e *PolicyInUseError) Error() string {
	return "policy is in use"
}

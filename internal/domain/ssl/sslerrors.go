package ssl

type SSLInUseError struct {
	Webapps []string
}

func (e *SSLInUseError) Error() string {
	return "SSL is in use"
}

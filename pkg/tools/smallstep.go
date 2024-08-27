package tools

import (
	"github.com/smallstep/certificates/cas/apiv1"
	"github.com/smallstep/certificates/pki"
)

// ref: https://github.com/smallstep/cli/blob/master/command/ca/init.go

// NewValues returns a new PKI configuration for the Smallstep Certificate Authority
func NewValues() (*pki.PKI, error) {
	p, err := pki.New(apiv1.Options{})
	if err != nil {
		return nil, err
	}

	return p, nil
}

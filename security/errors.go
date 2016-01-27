package security

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAppID = errors.New("invalid APP_ID")
)

type ErrPolicyNotFound struct {
	// type of policy, e.g. template or cap
	PolType string
	// apparmor or seccomp
	PolKind *securityPolicyType
	// name of the policy
	PolName string
}

func (e *ErrPolicyNotFound) Error() string {
	return fmt.Sprintf("could not find specified %s: %s (%s)", e.PolType, e.PolName, e.PolKind)
}

package service

import (
	"fmt"
	"strings"
)

type UserProvisionData struct {
	Email   string
	Subject string
}

func (u *UserProvisionData) CacheKey() string {
	e := strings.ToLower(u.Email)
	s := strings.ToLower(u.Subject)
	return fmt.Sprintf("%s:%s", s, e)
}

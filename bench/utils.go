package benchEtcd

import (
	up "github.com/satori/go.uuid"
)

func uuid() string {
	u, err := up.NewV4()
	if err != nil {
		panic(err)
	}
	return u.String()
}

package service

import "fmt"

type ObjectNotFoundError struct {
	AppId      int
	VersionId  int
	RevisionId int
	ObjectName string
	Name       string
}

func (err *ObjectNotFoundError) Error() string {
	return fmt.Sprintf("object not found: %+v", *err)
}

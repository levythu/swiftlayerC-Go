package outapi

// implement an empty io that does nothing

import (
    "errors"
    . "kernel/distributedvc/filemeta"
    "kernel/filetype"
    "io"
)

var (
    NOT_IMPLEMENTED_ERROR=errors.New("The method has not been implemented by purpose.")
)

type EmptyIO struct {
    // nothing
}

func (this *EmptyIO)GenerateUniqueID() string {
    return "outapi.EmptyIO"
}

// Need not have timestamp in FileMeta. It will be set according to content's record automatically.
// So do typestamp.
// Filemeta could be nil.
func (this *EmptyIO)Put(filename string, content filetype.Filetype, info FileMeta) error {
    return NOT_IMPLEMENTED_ERROR
}

// If file does not exist, a nil will be returned. No error occurs.
func (this *EmptyIO)Get(filename string) (FileMeta, filetype.Filetype, error) {
    return nil, nil, NOT_IMPLEMENTED_ERROR
}

func (this *EmptyIO)Putinfo(filename string, info FileMeta) error {
    return NOT_IMPLEMENTED_ERROR
}

// If file does not exist, a nil will be returned. No error occurs.
func (this *EmptyIO)Getinfo(filename string) (FileMeta, error) {
    return nil, NOT_IMPLEMENTED_ERROR
}

func (this *EmptyIO)Delete(filename string) error {
    return NOT_IMPLEMENTED_ERROR
}

// If file does not exist, a nil will be returned. No error occurs.
// Pay attention that io.ReadCloser should be closed.
func (this *EmptyIO)GetStream(filename string) (FileMeta, io.ReadCloser, error) {
    return nil, nil, NOT_IMPLEMENTED_ERROR
}

func (this *EmptyIO)PutStream(filename string, info FileMeta) (io.WriteCloser, error) {
    return nil, NOT_IMPLEMENTED_ERROR
}

// If the space is not available, create it and return (TRUE, nil);
// If the space is already available, return (FALSE, nil);
// Otherwise, (space is not available and fail to create), return a non-nil error.
func (this *EmptyIO)EnsureSpace() (bool, error) {
    return false, NOT_IMPLEMENTED_ERROR
}
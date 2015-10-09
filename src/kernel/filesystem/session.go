package filesystem
// A wrapper for fsfunc recording the working directory.

import (
    "outapi"
    "sync"
)

type Session struct {
    fs *Fs
    d string
    locks []*sync.Mutex
}
func NewSession(io outapi.Outapi) *Session {
    return &Session{
        fs: NewFs(io),
        d: ROOT_INODE_NAME,
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{}},
    }
}

func (this *Session)Cd(path string) error {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    var err error
    this.d, err=this.fs.Locate(path, this.d)
    return err
}

func (this *Session)Ls() ([]string, error) {
    return this.fs.List(this.d)
}

func (this *Session)Mkdir(foldername string) error {
    return this.fs.Mkdir(foldername, this.d)
}

func (this *Session)Rm(foldername string) error {
    return this.fs.Rm(foldername, this.d)
}

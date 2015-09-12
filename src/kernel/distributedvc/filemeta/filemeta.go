package filemeta

type FileMeta map[string]string

func NewMeta() FileMeta {
    return FileMeta(map[string]string{})
}

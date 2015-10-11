package filetype

// Kind of new filetype. Types like blob will implement both PointerType and FileType.
// For PointerType, any non-stream operation (Init and WriteBack) will not read the file itself, instead
// modifying the pointer value stored in its meta.

// Stream operation will lead actual io of the file content. But it's none of bussiness
// of the class.

// For implemented Filetype, its IsPointer() must return TRUE.

type PointerType interface {
    SetPointer(val string)
    GetPointer() string
}

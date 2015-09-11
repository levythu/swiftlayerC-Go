// K-V map file, storing string-to-string map based on the diction sort of keys,
// kernel file type to store folder index.
// File structure:
//     - BYTE0~3: magic chars "KVMP"
//     =========Rep=========
//     - 4B-int: n, 4B-int:m
//     - 8B-int: timestamp
//     - n B: unicoded key
//     - m B: unicoded value
//     =====================
//     - 0
// Store structure per entry:
//     (key,(value,timestamp))
// How to edit:
//     1. Construct with a string(create)/tuple(modify) as 2nd parameter
//     2. checkOut
//     3. edit kvm
//     4. checkIn
//     5. writeBack
// How to merge:
//     1. Construct with a tuple as 2nd parameter
//     2. mergeWith
//     3. [IF needs modification, checkIn/Out]
//     4. writeBack

package filetype

import (
    . "utils/timestamp"
    "io"
    "errors"
    "definition/exception"
    "encoding/binary"
    "log"
    "sort"
    "reflect"
    "fmt"
)

const fileMagic="KVMP"
const REMOVE_SPECIFIED="$@REMOVED@$)*!*"

type KvmapEntry struct {
    timestamp ClxTimestamp
    key string
    val string
}

type Kvmap struct {
    finishRead bool

    Kvm map[string]*KvmapEntry
    rmed map[string]*KvmapEntry

    readData []*KvmapEntry
    dataSource io.Reader

    fileTS ClxTimestamp
}

func Kvmap_verbose() {
    // NOUSE, only for crunching the warning.
    fmt.Print("useless")
}
func NewKvMap() *Kvmap {
    var nkv Kvmap
    rkv:=&nkv

    rkv.Init(nil, GetTimestamp(0))
    rkv.finishRead=true
    return rkv
}
func (this *Kvmap)Init(dtSource io.Reader, dtTimestamp ClxTimestamp) {
    this.readData=make([]*KvmapEntry, 0)
    this.dataSource=dtSource
    this.fileTS=dtTimestamp
    this.finishRead=false
}
func (this *Kvmap)GetType() string {
    return "key-value map file"
}

func ParseString(inp io.Reader ,length uint32) (string, error) {
    buf:=make([]byte, length)
    n, err:=inp.Read(buf)
    if err!=nil || uint32(n)<length {
        return "", errors.New(exception.EX_IMPROPER_DATA)
    }
    return string(buf[:n]), nil
}

func (this *Kvmap)GetTS() ClxTimestamp {
    return this.fileTS
}
func (this *Kvmap)SetTS(val ClxTimestamp) {
    this.fileTS=val
}
func (this *Kvmap)CheckOut() {
    // Attentez: All the modification will not be stored before executing CheckIn
    if this.LoadIntoMem()!=nil {
        return
    }
    this.Kvm=make(map[string]*KvmapEntry)
    this.rmed=make(map[string]*KvmapEntry)
    for _, elem:=range this.readData {
        if elem.val==REMOVE_SPECIFIED {
            this.rmed[elem.key]=elem
        } else {
            this.Kvm[elem.key]=elem
        }
    }
}
func (this *Kvmap)CheckIn() {
    if this.Kvm==nil {
        log.Fatal("<Kvmap::CheckIn> Have not checkout yet.")
    }
    tRes:=make([]*KvmapEntry, 0)
    keyArray:=make([]string, 0)

    for key:=range this.Kvm {
        keyArray=append(keyArray,key)
    }
    for key:=range this.rmed {
        if _, ok:=this.Kvm[key]; !ok {
            keyArray=append(keyArray,key)
        }
    }
    sort.Strings(keyArray)

    for _, key:=range keyArray {
        val4kvm, ok4kvm:=this.Kvm[key]
        val4rm, ok4rm:=this.rmed[key]
        if ok4kvm && ok4rm {
            if val4kvm.timestamp<val4rm.timestamp {
                tRes=append(tRes, val4rm)
            } else {
                tRes=append(tRes, val4kvm)
            }
        }
        if ok4kvm && !ok4rm {
            tRes=append(tRes, val4kvm)
        }
        if !ok4kvm && ok4rm {
            tRes=append(tRes, val4rm)
        }
    }

    this.readData=tRes
}

func (this *Kvmap)MergeWith(file2 Filetype) error {
    if reflect.TypeOf(this)!=reflect.TypeOf(file2) {
        return errors.New(exception.EX_UNMATCHED_MERGE)
    }
    tRes:=make([]*KvmapEntry, 0)
    file2x:=file2.(*Kvmap)
    i,j:=0,0

    for {
        if this.lazyRead_NoError(i)==nil {
            for file2x.lazyRead_NoError(j)!=nil {
                tRes=append(tRes,file2x.lazyRead_NoError(j))
                j=j+1
            }
            break
        }
        if file2x.lazyRead_NoError(j)==nil {
            for this.lazyRead_NoError(i)!=nil {
                tRes=append(tRes,this.lazyRead_NoError(i))
                i=i+1
            }
            break
        }
        for this.lazyRead_NoError(i)!=nil && file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i).key<file2x.lazyRead_NoError(j).key {
            tRes=append(tRes,this.lazyRead_NoError(i))
            i=i+1
        }
        for file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i)!=nil && this.lazyRead_NoError(i).key>file2x.lazyRead_NoError(j).key {
            tRes=append(tRes,file2x.lazyRead_NoError(j))
            j=j+1
        }
        for file2x.lazyRead_NoError(j)!=nil && this.lazyRead_NoError(i)!=nil && this.lazyRead_NoError(i).key==file2x.lazyRead_NoError(j).key {
            if this.lazyRead_NoError(i).timestamp>file2x.lazyRead_NoError(j).timestamp {
                tRes=append(tRes,this.lazyRead_NoError(i))
            } else if this.lazyRead_NoError(i).timestamp<file2x.lazyRead_NoError(j).timestamp {
                tRes=append(tRes,file2x.lazyRead_NoError(j))
            } else {
                // Attentez: this conflict resolving strategy may be altered.
                tRes=append(tRes,this.lazyRead_NoError(i))
            }
            i=i+1
            j=j+1
        }
    }
    this.readData=tRes
    this.fileTS=MergeTimestamp(this.fileTS,file2x.fileTS)

    return nil
}

func (this *Kvmap)WriteBack(dtDes io.Writer) error {
    if err:=this.LoadIntoMem(); err!=nil {
        return err
    }
    if _,err:=dtDes.Write([]byte(fileMagic)); err!=nil {
        return err
    }
    for _, elem:=range this.readData {
        K:=[]byte(elem.key)
        V:=[]byte(elem.val)
        if err:=binary.Write(dtDes, binary.LittleEndian, uint32(len(K))); err!=nil {
            return err
        }
        if err:=binary.Write(dtDes, binary.LittleEndian, uint32(len(V))); err!=nil {
            return err
        }
        if err:=binary.Write(dtDes, binary.LittleEndian, elem.timestamp); err!=nil {
            return err
        }
        if _,err:=dtDes.Write(K); err!=nil {
            return err
        }
        if _,err:=dtDes.Write(V); err!=nil {
            return err
        }
    }
    if err:=binary.Write(dtDes, binary.LittleEndian, uint32(0)); err!=nil {
        return err
    }
    return nil
}
func (this *Kvmap)LoadIntoMem() error {
    for !this.finishRead {
        _, err:=this.lazyRead(len(this.readData))
        if err!=nil {
            return err
        }
    }
    return nil
}
func (this *Kvmap)lazyRead_NoError(pos int) *KvmapEntry {
    res, err:=this.lazyRead(pos)
    if err!=nil {
        return nil
    }
    return res
}
func (this *Kvmap)lazyRead(pos int) (*KvmapEntry, error) {
    if pos<len(this.readData) {
        return this.readData[pos], nil
    }
    if this.finishRead {
        return nil, nil
    }
    if len(this.readData)==0 {
        // Open the target, check it.
        tmpString, err:=ParseString(this.dataSource, 4)
        if (err!=nil) {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }
        if tmpString!=fileMagic {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }
    }

    for pos>=len(this.readData) {
        var m, n uint32
        var ts ClxTimestamp
        if binary.Read(this.dataSource, binary.LittleEndian, &n)!=nil {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }
        if n==0 {
            this.finishRead=true
            return nil, nil
        }
        if binary.Read(this.dataSource, binary.LittleEndian, &m)!=nil {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }
        if binary.Read(this.dataSource, binary.LittleEndian, &ts)!=nil {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }

        K, err:=ParseString(this.dataSource, n)
        if (err!=nil) {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }

        V, err:=ParseString(this.dataSource, m)
        if (err!=nil) {
            return nil, errors.New(exception.EX_WRONG_FILEFORMAT)
        }

        this.readData=append(this.readData, &KvmapEntry{
            timestamp: ts,
            key: K,
            val: V,
        })
    }

    return this.readData[pos], nil
}

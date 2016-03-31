package intranet

import (
    . "github.com/levythu/gurgling"
    "github.com/levythu/gurgling/midwares/auth"
    "github.com/levythu/gurgling/midwares/staticfs"
    conf "definition/configinfo"
    . "logger"
    "sync"
    . "definition"
    "time"
    dvc "kernel/distributedvc"
)

func getAdminPageRouter() Router {
    var r=ARouter()

    if conf.ADMIN_USER!="" {
        r.Use(auth.ABasicAuth(conf.ADMIN_USER, conf.ADMIN_PASSWORD, ":[intranet]/admin"))
    } else {
        Secretary.Warn("intranet::getAdminPageRouter()", "Administrator authentication is canceled. Please ensure the inner service is "+
            "running on a safe network, otherwise set inner_service_admin_user in cofiguration.")
    }
    var p, err=GetABSPath("./public/intranet")
    if err!=nil {
        Secretary.Error("intranet::getAdminPageRouter()", "Fail to locate public directory. Intranet service stops.")
        return nil
    }
    r.Use("/", staticfs.AStaticfs(p))
    r.Get("/taskinfo", getMergingTaskInfo)
    r.Get("/loginfo", getLoggingInfo)
    r.Get("/fdinfo", getFDInfo)

    return r
}


var gMTI_recordTime int64=0
var gMTI_Cache map[string]interface{}
var gMTI_lock=sync.RWMutex{}
func getMergingTaskInfo(req Request, res Response) {
    var nTime=time.Now().Unix()
    gMTI_lock.RLock()
    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gMTI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gMTI_recordTime,
            "val":          gMTI_Cache,
        })
        gMTI_lock.RUnlock()
        return
    }
    gMTI_lock.RUnlock()
    gMTI_lock.Lock()
    defer gMTI_lock.Unlock()

    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gMTI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gMTI_recordTime,
            "val":          gMTI_Cache,
        })
        return
    }


    gMTI_Cache=map[string]interface{} {
        "worksAlive":   dvc.MergeManager.Reveal_workersAlive(),
        "taskInfo":     dvc.MergeManager.Reveal_taskInfo(),
    }
    gMTI_recordTime=time.Now().Unix()


    res.JSON(map[string]interface{}{
        "recordsTime":  gMTI_recordTime,
        "val":          gMTI_Cache,
    })
}



var gLI_recordTime int64=0
var gLI_Cache map[string]interface{}
var gLI_lock=sync.RWMutex{}
func getLoggingInfo(req Request, res Response) {
    var nTime=time.Now().Unix()
    gLI_lock.RLock()
    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gLI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gLI_recordTime,
            "val":          gLI_Cache,
        })
        gLI_lock.RUnlock()
        return
    }
    gLI_lock.RUnlock()
    gLI_lock.Lock()
    defer gLI_lock.Unlock()

    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gLI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gLI_recordTime,
            "val":          gLI_Cache,
        })
        return
    }

    var tmpList=[]interface{}{}
    var dRes=true
    if _, ok:=req.Query()["log"]; ok {
        dRes=dRes&&SecretaryCache.Dump(func(obj CachedLoggerEntry) bool {
            tmpList=append(tmpList, map[string]interface{} {
                "pos":      obj.Pos,
                "content":  obj.Content,
                "time":     obj.Time.UnixNano(),
                "type":     "log",
            })
            return true
        }, 0)
    }
    if _, ok:=req.Query()["warn"]; ok {
        dRes=dRes&&SecretaryCache.Dump(func(obj CachedLoggerEntry) bool {
            tmpList=append(tmpList, map[string]interface{} {
                "pos":      obj.Pos,
                "content":  obj.Content,
                "time":     obj.Time.UnixNano(),
                "type":     "warn",
            })
            return true
        }, 1)
    }
    if _, ok:=req.Query()["error"]; ok {
        dRes=dRes&&SecretaryCache.Dump(func(obj CachedLoggerEntry) bool {
            tmpList=append(tmpList, map[string]interface{} {
                "pos":      obj.Pos,
                "content":  obj.Content,
                "time":     obj.Time.UnixNano(),
                "type":     "error",
            })
            return true
        }, 2)
    }

    gLI_Cache=map[string]interface{} {
        "available":    dRes,
        "loglist":      tmpList,
    }
    gLI_recordTime=time.Now().Unix()


    res.JSON(map[string]interface{}{
        "recordsTime":  gLI_recordTime,
        "val":          gLI_Cache,
    })
}



var gFDI_recordTime int64=0
var gFDI_Cache map[string]interface{}
var gFDI_lock=sync.RWMutex{}
func getFDInfo(req Request, res Response) {
    var nTime=time.Now().Unix()
    gFDI_lock.RLock()
    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gFDI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gFDI_recordTime,
            "val":          gFDI_Cache,
        })
        gFDI_lock.RUnlock()
        return
    }
    gFDI_lock.RUnlock()
    gFDI_lock.Lock()
    defer gFDI_lock.Unlock()

    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gFDI_recordTime {
        res.JSON(map[string]interface{}{
            "recordsTime":  gFDI_recordTime,
            "val":          gFDI_Cache,
        })
        return
    }

    gFDI_Cache=dvc.Reveal_FdPoolProfile()
    gFDI_recordTime=time.Now().Unix()


    res.JSON(map[string]interface{}{
        "recordsTime":  gFDI_recordTime,
        "val":          gFDI_Cache,
    })
}
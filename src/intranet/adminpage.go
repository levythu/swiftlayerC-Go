package intranet

import (
    . "github.com/levythu/gurgling"
    "github.com/levythu/gurgling/midwares/auth"
    "github.com/levythu/gurgling/midwares/staticfs"
    conf "definition/configinfo"
    . "logger"
    "sync"
    "time"
)

var minFreq=time.Duration(conf.ADMIN_REFRESH_FREQUENCY)*time.Second

func getAdminPageRouter() Router {
    var r=ARouter()

    if conf.ADMIN_USER!="" {
        r.Use(auth.ABasicAuth(conf.ADMIN_USER, conf.ADMIN_PASSWORD, ":[intranet]/admin"))
    } else {
        Secretary.Warn("Administrator authentication is canceled. Please ensure the inner service is "+
            "running on a safe network, otherwise set inner_service_admin_user in cofiguration.")
    }
    r.Use("/", staticfs.AStaticfs("./public/intranet"))
    r.Get("/TaskInfo", getMergingTaskInfo)
}

var gMTI_recordTime=0
var gMTI_Cache=""
var gMTI_lock=sync.RWMutex{}
func getMergingTaskInfo(req Request, res Response) {
    var nTime=time.Now().Unix()
    gMTI_lock.RLock()
    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gMTI_recordTime {
        res.JSON(map[string]string{
            "recordsTime": gMTI_recordTime,
            "val": gMTI_Cache,      // TODO: escape gMTI_Cache
        })
        gMTI_lock.RUnlock()
        return
    }
    gMTI_lock.RUnlock()
    gMTI_lock.Lock()
    defer gMTI_lock.Unlock()

    if nTime<conf.ADMIN_REFRESH_FREQUENCY+gMTI_recordTime {
        res.JSON(map[string]string{
            "recordsTime": gMTI_recordTime,
            "val": gMTI_Cache,      // TODO: escape gMTI_Cache
        })
        return
    }

    
}

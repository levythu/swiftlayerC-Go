package inapi

import (
    . "github.com/levythu/gurgling"
    "inapi/containermanage"
    "inapi/fsmanage"
    "inapi/streamio"
    conf "definition/configinfo"
)

func Entry() {
    var rootRouter=ARouter()
    rootRouter.Use("/fs", fsmanage.FMRouter())
    rootRouter.Use("/io", streamio.IORouter())
    rootRouter.Use("/cn", containermanage.CMRouter())

    rootRouter.Launch(conf.OUTER_SERVICE_LISTENER)
}

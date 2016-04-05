package gossipd

import (
    . "github.com/levythu/gurgling"
    "io/ioutil"
    "outapi"
    dvc "kernel/distributedvc"
    . "logger"
    . "intranet/gossipd/interactive"
)

// @ async
func checkGossipedData(src []*GossipEntry) {
    for _, e:=range src {
        if io:=outapi.DeSerializeID(e.OutAPI); io==nil {
            Secretary.Warn("gossipd::checkGossipedData()", "Invalid Outapi DeSerializing: "+e.OutAPI)
            continue
        } else {
            if fd:=dvc.GetFD(e.Filename, io); fd==nil {
                Secretary.Warn("gossipd::checkGossipedData()", "Fail to get FD for "+e.Filename)
                continue
            } else {
                // TODO!
                fd.Release()
            }
        }
    }
}
/*
** GOSSIP API: Posted
** Method:      POST
** URL:         [:intranet]/gossip
** Parameter:   Content(in Body): the raw body is the parameter content itself.
*/
func OnPostedGossip(req Request, res Response) {
    if ct, err:=ioutil.ReadAll(req.R().Body); err!=nil {
        Secretary.Error("gossipd::OnPostedGossip", "Fail to read data from gossiped request: "+err.Error())
        res.SendCode(500)
        return
    } else {
        var pList=ParseAll(string(ct))
        if pList==nil {
            Secretary.Error("gossipd::OnPostedGossip", "Format error for gossiped data")
            res.SendCode(403)
            return
        }

        go checkGossipedData(pList)
        res.SendCode(200)
    }
}

func GetGossipRouter() Router {
    var r=ARouter()
    r.Post("/", OnPostedGossip)

    return r
}

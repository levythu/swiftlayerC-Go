package gossipd

import (
    . "github.com/levythu/gurgling"
    "io/ioutil"
    "outapi"
    dvc "kernel/distributedvc"
    . "logger"
    gsp "intranet/gossip"
    . "intranet/gossipd/interactive"
)

// @ async
func checkGossipedData(src []*GossipEntry) {
    // TODO: whether use multi-routine?
    for _, e:=range src {
        if io:=outapi.DeSerializeID(e.OutAPI); io==nil {
            Secretary.Warn("gossipd::checkGossipedData()", "Invalid Outapi DeSerializing: "+e.OutAPI)
            continue
        } else {
            if fd:=dvc.GetFD(e.Filename, io); fd==nil {
                Secretary.Warn("gossipd::checkGossipedData()", "Fail to get FD for "+e.Filename)
                continue
            } else {
                fd.GraspReader()
                if !fd.ASYNCMergeWithNodeX(e) {
                    // the fd need not gossiped. SO just propagate the original one
                    if err:=gsp.GlobalGossiper.PostGossip(e); err!=nil {
                        Secretary.Warn("gossipd::checkGossipedData", "Fail to post change gossiping to other nodes: "+err.Error())
                    }
                } else {
                    // the file itself needs gossiping. wait for it to writeback and trigger gossiping
                    // DO NOTHING now
                }
                fd.ReleaseReader()
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

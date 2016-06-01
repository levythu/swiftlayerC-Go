var h2=require("./h2");
var fs=require("fs");

var args=process.argv;
var isRaw=false;
var fList=JSON.parse(fs.readFileSync("./pattern.txt", {encoding: "utf8"}));
var container="defaultCon";
if (args.length>2) {
    container=args[2];
}
if (args.length>3) {
    isRaw=true;
}

var globalTime=0;
(function goList(i, callback) {
    if (i==fList.length) {
        callback();
        return;
    }
    var elem=fList[i];
    h2.getFile(container, elem, function(time) {
        console.log("Fetched", elem, "in", time, "ms");
        globalTime+=time;
        process.nextTick(function() {
            goList(i+1, callback);
        });
    }, isRaw);

})(0, function() {
    console.log("average fetch time:", globalTime/fList.length)
    process.exit(0);
});

var map;
const centerLatitude=35.6581;
const centerLongitude=139.6975;
var shapes;

var addShape = function(shape){
  shape.setMap(map);
  var now = Date.now()
  shapes.push({validThru:now+5*1000, shape:shape});
  while( shapes[0].validThru < now ){
    let v = shapes[0]
    console.log("delete:"+v.validThru)
    v.shape.setMap(null)
    shapes.shift()
  }
}

var drawMap = function(lat,lng){
  map = new google.maps.Map(document.getElementById('map'), {
    center: {lat: lat, lng: lng},
    mapTypeControl: false,
    zoom: 18
  });
  // Google Mapで情報ウインドウの表示、非表示の切り替え
  // https://hacknote.jp/archives/19977/
  (function fixInfoWindow() {
  var set = google.maps.InfoWindow.prototype.set;
  google.maps.InfoWindow.prototype.set = function(key, val) {
    if (key === "map") {
      if (! this.get("noSuppress")) {
        return;
      }
    }
    set.apply(this, arguments);
  }
  }());
}

//var doPost = function(jsonArray){
var doPost = function(url,jsonArray,handleFunc){
  //console.log('doPost:'+jsonArray.length);
  //if (jsonArray.length==0) return;
  var req = new XMLHttpRequest();
  var callback = function(response){
    let result = JSON.parse(response);
    //console.log("result="+result)
    //console.log("result.length="+result.length)
    for (i=0;i<result.length;i++){
      let way = result[i]
      // Skip building
      if (way.Tags.building ) continue;

      let lines =[];
      for (j=0;j<way.Nodes.length;j++){
        let node = way.Nodes[j]
        if(node){
          lines.push({lat: node.Lat, lng:node.Lon})
          var circle = new google.maps.Circle({
            strokeColor: '#0000FF',
            strokeOpacity: 0.8,
            strokeWeight: 1,
            fillColor: '#0000FF',
            fillOpacity: 0.35,
            center: {lat: node.Lat, lng:node.Lon},
            radius: 2
          });
          circle.setMap(map);

          circle.addListener('click', function(e) {
            console.log("Node:===============");
            for(prop in node.Tags){
              console.log(" "+prop+": "+node.Tags[prop]);
            }
          });
        }
      }

      let color;
      if ( isFootway(way.Tags) ){
        color = '#0000FF'
      } else {
        color = '#00FFFF'
      }
      var polyline = new google.maps.Polyline({
        path: lines,
        geodesic: true,
        strokeColor: color,
        strokeOpacity: 1.0,
        strokeWeight: 3
        });
      //addShape(polyline)
      polyline.setMap(map);

      polyline.addListener('click', function(e) {
        console.log("Way:===============");
        way.strokeWeight=5
        for(prop in way.Tags){
          console.log(" "+prop+": "+way.Tags[prop]);
        }
      });
    }
  }
  req.onreadystatechange = function() {
    if (req.readyState == 4) { // finished sending
      console.log("req.status="+req.status);
      if (req.status == 200) {
        //console.log(req.responseText);
        //callback(req.responseText);
        handleFunc(req.responseText);
      }
    }else{
      console.log("通信中...");
    }
  }
  req.open('POST', url, true);
  req.setRequestHeader("Content-type", "application/json");
  var parameters = JSON.stringify(jsonArray);
  req.send(parameters);
}

function geoInfo() {
  this.json = [];
  this.postTimer = 0;
  this.Interval = 5000; // 5 seconds
  //this.Interval = 60000; // 60 seconds
};
geoInfo.prototype = {
  //json      : [] ,
  clearJson : function() {
    this.json=[];
  },
  pushJson  : function(id,time,lat,lng){
    this.json.push({
      "consumerId"  : id ,
      "timestamp"  : time ,
      "latitude"  : lat ,
      "longtitude"  : lng
    });
    //console.log("pushJson:"+this.json.length+" :"+this.json[this.json.length-1]);
  },
  //postTimer     : 0,
  stopPostTimer : function() {
    clearTimeout(this.postTimer);
    this.postTimer=0;
  },
  startPost: function(){
    this.stopPostTimer(); // avoid duplicate timer
    this.postTimer=setTimeout(this.post.bind(this), this.Interval);
  },
  post          : function() {
    if (this.json.length>0){
      doPost('/location',this.json,function(){});
      this.clearJson();
    };
    this.postTimer=setTimeout(this.post.bind(this), this.Interval);
  }
}

var initMap = function() {

  var info = new geoInfo();
  shapes=[]
  console.log('Lat=' + centerLatitude + ' Lng=' + centerLongitude);
  drawMap(centerLatitude,centerLongitude);

  // Update lat/long value of div when anywhere in the map is clicked
  google.maps.event.addListener(map,'click',function(event) {
    var lat = event.latLng.lat();
    var lng = event.latLng.lng();
    //document.getElementById('currentLat').innerHTML = lat;
    //document.getElementById('currentLon').innerHTML = lng;
    info.pushJson(1 ,new Date() , lat , lng);
    var circle = new google.maps.Circle({
      strokeColor: '#FF0000',
      strokeOpacity: 0.8,
      strokeWeight: 1,
      fillColor: '#FF0000',
      fillOpacity: 0.35,
      //map: map,
      center: {lat: lat, lng:lng},
      radius: 2
    });
    addShape(circle);
  });
  info.startPost();
};


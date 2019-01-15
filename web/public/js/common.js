var map;
const centerLatitude=35.6581;
const centerLongitude=139.6975;
var shapes;

var addShape = function(shape){
  shape.setMap(map);
  var now = Date.now()
  shapes.push({validThru:now+0.9*1000, shape:shape});
  while( shapes[0].validThru < now ){
    let v = shapes[0]
    //console.log("delete:"+v.validThru)
    v.shape.setMap(null)
    shapes.shift()
  }
}

var drawMap = function(lat,lng){
  map = new google.maps.Map(document.getElementById('map'), {
    center: {lat: lat, lng: lng},
    mapTypeControl: false,
    zoomControl: false,
    streetViewControl: false,
    zoom: 17
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

var doPost = function(url,requestInfo,handleFunc){
  //console.log('doPost:'+requestInfo.length);
  var req = new XMLHttpRequest();
  req.onreadystatechange = function() {
    if (req.readyState == 4) { // finished sending
      //console.log("req.status="+req.status);
      if (req.status == 200) {
        //console.log(req.responseText);
        handleFunc(req.responseText);
      }
    }else{
      //console.log("通信中...");
    }
  }
  req.open('POST', url, true);
  req.setRequestHeader("Content-type", "application/json");
  var parameters = JSON.stringify(requestInfo);
  req.send(parameters);
}

function geoInfo() {
  this.request = {
      bounds: {},
      flyers: [],
  };
  this.postTimer = 0;
  this.Interval = 1000; // 1 seconds
  //this.Interval = 60000; // 60 seconds
};

geoInfo.prototype = {
  clearRequest : function() {
    this.request={
      bounds: {},
      flyers: [],
    };
  },
  setBounds  : function(bounds){
    this.request.bounds=bounds;
    //console.log("pushFlyer:"+this.request.length+" :"+this.request[this.request.length-1]);
  },
  pushFlyer  : function(id,time,lat,lng,title,distance){
    this.request.flyers.push({
      "storeId"     : id,
      "title"       : title,
      "timestamp"   : time,
      "latitude"    : lat,
      "longitude"   : lng,
      "distance"    : distance,
      "pieces"      : 10,
    });
    //console.log("pushFlyer:"+this.request.length+" :"+this.request[this.request.length-1]);
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
  drawLocations: function(response){
    //console.log(response);
    let locations = JSON.parse(response);
    //console.log("LOCATIONS COUNT:"+locations.length);
    for(i=0;i<locations.length;i++){
      let loc=locations[i]
      //console.log(loc);
      var circle = new google.maps.Circle({
        strokeColor: '#000000',
        strokeOpacity: 1.0,
        strokeWeight: 1,
        fillColor: '#020202',
        fillOpacity: 0.8,
        //map: map,
        center: {lat: loc.geometry.coordinates[1], lng:loc.geometry.coordinates[0]},
        radius: 2,
      });
      addShape(circle);

      /*
      var marker = new google.maps.Marker({
        position: {lat: loc.geometry.coordinates[1], lng:loc.geometry.coordinates[0]},
        flat: true,
        title: "marker title!!",
        cursor: "marker cursor!?",
        label: String(5),
        //icon: google.maps.SymbolPath.CIRCLE, // error
      });
      addShape(marker);
      */
    }
  },
  post          : function() {
    this.setBounds(map.getBounds());
    doPost('/location',this.request,this.drawLocations);
    this.clearRequest();
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
    var title = String(document.forms.form1.title.value);
    var distance = Number(document.forms.form1.distance.value);
    console.log("title="+title)
    console.log("distance="+distance)
    info.pushFlyer(1 ,new Date() , lat, lng, title, distance);
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


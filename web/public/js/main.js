var map;
const centerLatitude=35.6581;
const centerLongitude=139.6975;
var shapes;
var notifIDs; // Already notified notification IDs
var notifIDsByUserID;

var locInfo;

var flyerInfo;

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
    zoom: 15
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

var doHttp = function(method,url,requestInfo,handleFunc){
  //console.log('doHttp:'+requestInfo.length);
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
  req.open(method, url, true);
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
  pushFlyer  : function(id,validPeriod,lat,lng,title,distance){
    this.request.flyers.push({
      "storeId"     : id,
      "title"       : title,
      "validPeriod" : validPeriod,
      "latitude"    : lat,
      "longitude"   : lng,
      "distance"    : distance,
      "stocked"     : 10,
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
  drawResponse: function(responseJson){
    //console.log(responseJson);
    let response = JSON.parse(responseJson);
    //let locations = response.locations;
    //let flyers = response.flyers;
    let notifs = response.notifications;
    //console.log("FLYERS COUNT:"+flyers.length);
    console.log("notifs:"+notifs.length);

    var userIDsToNotif={}
    for(i=0;i<notifs.length;i++){
      notif=notifs[i];
      if ( !notifIDs[notif.properties.id] && !locInfo.userLocs[notif.properties.userId]){
        //console.log("No userLocs userId:"+notif.properties.userId);
      }
      if ( !notifIDs[notif.properties.id] && locInfo.userLocs[notif.properties.userId]){
        //console.log(notif);
        console.log("New notification ID:"+notif.properties.id+" UserID:"+notif.properties.userId);
        notifIDs[notif.properties.id]=true;
        if (! userIDsToNotif[notif.properties.userId]) {
          userIDsToNotif[notif.properties.userId]=[notif.properties.id];
        } else {
          userIDsToNotif[notif.properties.userId].push(notif.properties.id);
        }
        if (! notifIDsByUserID[notif.properties.userId]) {
          notifIDsByUserID[notif.properties.userId]=[notif.properties.id];
        } else {
          notifIDsByUserID[notif.properties.userId].push(notif.properties.id);
        }
      }
    }

    for(userId in userIDsToNotif){
      let marker = new google.maps.Marker({
        position: {lat: locInfo.userLocs[userId].lat, lng: locInfo.userLocs[userId].lng},
        flat: true,
        title: "marker title!!",
        cursor: "marker cursor!?",
        //label: String(userIDsToNotif[userId].length),
        label: String(notifIDsByUserID[userId].length),
        //icon: google.maps.SymbolPath.CIRCLE, // error
      });
      marker.setMap(map);
      setTimeout(function(){
        marker.setMap(null) // Remove Marker
      }, 3*1000);
    }
  },
  post          : function() {
    this.setBounds(map.getBounds());
    doHttp('POST','/location',this.request,this.drawResponse);
    this.clearRequest();
    this.postTimer=setTimeout(this.post.bind(this), this.Interval);
  }
}

var initMap = function() {
  var info = new geoInfo();
  shapes=[];
  notifIDs={};
  notifIDsByUserID={};
  console.log('Lat=' + centerLatitude + ' Lng=' + centerLongitude);
  drawMap(centerLatitude,centerLongitude);

  // Update lat/long value of div when anywhere in the map is clicked
  google.maps.event.addListener(map,'click',function(event) {
    var lat = event.latLng.lat();
    var lng = event.latLng.lng();
    var title = String(document.forms.form1.title.value);
    var distance = Number(document.forms.form1.distance.value);
    var validPeriod = Number(document.forms.form1.validPeriod.value);
    console.log("title="+title)
    console.log("distance="+distance)
    info.pushFlyer(1 ,validPeriod , lat, lng, title, distance);
    let circle = new google.maps.Circle({
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

var initHttp = function() {
  locInfo = new locationInfo();
  locInfo.startPost()
  fInfo = new flyerInfo();
  console.log(fInfo)
  fInfo.startPost()
}

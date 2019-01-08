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

var doPost = function(jsonArray){
	console.log('doPost:'+jsonArray.length);
	if (jsonArray.length==0) return;
	var req = new XMLHttpRequest();
	req.onreadystatechange = function() {
		if (req.readyState == 4) { // finished sending
			console.log("req.status="+req.status);
			if (req.status == 200) {
				console.log(req.responseText);
			}
		}else{
			console.log("通信中...");
		}
	}
	req.open('POST', '/location', true);
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
			"consumerId"	: id ,
			"timestamp"	: time ,
			"latitude"	: lat ,
			"longtitude"	: lng
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
		doPost(this.json);
		this.clearJson();
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


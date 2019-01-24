var map;
const centerLatitude=35.6581;
const centerLongitude=139.6975;
var shapes;
var locInfo;
var flyInfo;
//var notifInfo;

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

var initMap = function() {
	shapes=[];
	console.log('Lat=' + centerLatitude + ' Lng=' + centerLongitude);
	drawMap(centerLatitude,centerLongitude);

	// Update lat/long value of div when anywhere in the map is clicked
	google.maps.event.addListener(map,'click',function(event) {
		var lat = event.latLng.lat();
		var lng = event.latLng.lng();
		console.log("lat: "+lat+" , lng: "+lng);
		var title = String(document.forms.form1.title.value);
		var distance = Number(document.forms.form1.distance.value);
		var validPeriod = Number(document.forms.form1.validPeriod.value);
		console.log("title="+title)
		console.log("distance="+distance)
		flyInfo.post(1 ,validPeriod , lat, lng, title, distance);
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
};

var initHttp = function() {
	locInfo = new locationInfo();
	locInfo.startGet();

	flyInfo = new flyerInfo();
	flyInfo.startGet();

	notifInfo = new notifInfo();
	notifInfo.startGet();
}

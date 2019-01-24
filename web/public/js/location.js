function locationInfo() {
	this.postTimer = 0;
	this.Interval = 1000; // 1 seconds
	this.userLocs = {};
};

locationInfo.prototype = {
	stopGetTimer : function() {
		clearTimeout(this.postTimer);
		this.postTimer=0;
	},
	startGet: function(){
		this.stopGetTimer(); // avoid duplicate timer
		this.postTimer=setTimeout(this.post.bind(this), this.Interval);
	},
	drawResponse: function(responseJson){
		let locations = JSON.parse(responseJson);

		for(i=0;i<locations.length;i++){
			let loc=locations[i]
			//console.log(loc);
			let circle = new google.maps.Circle({
				strokeColor: '#000000',
				strokeOpacity: 1.0,
				strokeWeight: 1,
				fillColor: '#020202',
				fillOpacity: 0.8,
				//map: map,
				center: {lat: loc.geometry.coordinates[1], lng:loc.geometry.coordinates[0]},
				radius: 2,
			});
			this.userLocs[loc.properties.id]={lat: loc.geometry.coordinates[1], lng:loc.geometry.coordinates[0]};
			addShape(circle);
		}
		console.log("locations:"+locations.length+" userLocs:"+Object.keys(this.userLocs).length);
		//console.log(this.userLocs);
	},
	post          : function() {
		let bounds=map.getBounds();
		//console.log(bounds);
		let url='/api/locations?south='+bounds.getSouthWest().lat()+'&west='+bounds.getSouthWest().lng()+'&north='+bounds.getNorthEast().lat()+'&east='+bounds.getNorthEast().lng()
		doHttp('GET',url,this.request,this.drawResponse.bind(this));
		this.postTimer=setTimeout(this.post.bind(this), this.Interval);
	}
}

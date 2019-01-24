function notifInfo() {
	this.postTimer = 0;
	this.Interval = 1000; // 1 seconds
	this.notifIDs={}; // Already notified notification IDs
	this.notifIDsByUserID={};
};

notifInfo.prototype = {
	stopGetTimer : function() {
		clearTimeout(this.postTimer);
		this.postTimer=0;
	},
	startGet: function(){
		this.stopGetTimer(); // avoid duplicate timer
		this.postTimer=setTimeout(this.post.bind(this), this.Interval);
	},
	drawResponse: function(responseJson){
		let notifs = JSON.parse(responseJson);
		var userIDsToNotif={}
		for(i=0;i<notifs.length;i++){
			notif=notifs[i];
			if ( !this.notifIDs[notif.properties.id] && !locInfo.userLocs[notif.properties.userId]){
				//console.log("No userLocs userId:"+notif.properties.userId);
			}
			if ( !this.notifIDs[notif.properties.id] && locInfo.userLocs[notif.properties.userId]){
				//console.log(notif);
				console.log("New notification ID:"+notif.properties.id+" UserID:"+notif.properties.userId);
				this.notifIDs[notif.properties.id]=true;
				if (! userIDsToNotif[notif.properties.userId]) {
					userIDsToNotif[notif.properties.userId]=[notif.properties.id];
				} else {
					userIDsToNotif[notif.properties.userId].push(notif.properties.id);
				}
				if (! this.notifIDsByUserID[notif.properties.userId]) {
					this.notifIDsByUserID[notif.properties.userId]=[notif.properties.id];
				} else {
					this.notifIDsByUserID[notif.properties.userId].push(notif.properties.id);
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
				label: String(this.notifIDsByUserID[userId].length),
				//icon: google.maps.SymbolPath.CIRCLE, // error
			});
			marker.setMap(map);
			setTimeout(function(){
				marker.setMap(null) // Remove Marker
			}, 3*1000);
		}
	},
	post          : function() {
		let bounds=map.getBounds();
		let url='/api/notifications?south='+bounds.getSouthWest().lat()+'&west='+bounds.getSouthWest().lng()+'&north='+bounds.getNorthEast().lat()+'&east='+bounds.getNorthEast().lng()
		doHttp('GET',url,this.request,this.drawResponse.bind(this));
		this.postTimer=setTimeout(this.post.bind(this), this.Interval);
	}
}

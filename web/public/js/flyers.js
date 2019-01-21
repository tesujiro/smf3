function flyerInfo() {
  this.postTimer = 0;
  this.Interval = 1000; // 1 seconds
};

flyerInfo.prototype = {
  stopPostTimer : function() {
    clearTimeout(this.postTimer);
    this.postTimer=0;
  },
  startPost: function(){
    this.stopPostTimer(); // avoid duplicate timer
    this.postTimer=setTimeout(this.post.bind(this), this.Interval);
  },
  drawResponse: function(responseJson){
    let flyers = JSON.parse(responseJson);
    //console.log(flyers);
    let now=Math.floor((new Date).getTime()/1000);
    for(i=0;i<flyers.length;i++){
      let flyer=flyers[i]
      //console.log(flyer);
      //console.log("now:"+now);
      //console.log("start:"+flyer.properties.startAt)
      //console.log("end:"+flyer.properties.endAt)
      if ( !flyerIDs[flyer.properties.id] 
        && flyer.properties.startAt <= now
        && now <= flyer.properties.endAt
      ){
        //console.log("==>WRITE CIRCLE!")
        flyerIDs[flyer.properties.id]=true;
        let circle = new google.maps.Circle({
          strokeColor: '#80FF00',
          strokeOpacity: 0.6,
          strokeWeight: 0.4,
          fillColor: '#80FF00',
          fillOpacity: 0.3,
          center: {lat: flyer.geometry.coordinates[1], lng:flyer.geometry.coordinates[0]},
          radius: flyer.properties.distance,
        });
        circle.setMap(map);
        setTimeout(function(){
          circle.setMap(null) // Remove Circle
        }, (flyer.properties.endAt - now)*1000);
      }
    }
  },
  post          : function() {
    let bounds=map.getBounds();
    //console.log(bounds);
    let url='/api/flyers?south='+bounds.getSouthWest().lat()+'&west='+bounds.getSouthWest().lng()+'&north='+bounds.getNorthEast().lat()+'&east='+bounds.getNorthEast().lng()
    doHttp('GET',url,this.request,this.drawResponse);
    this.postTimer=setTimeout(this.post.bind(this), this.Interval);
  }
}

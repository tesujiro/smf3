function flyerInfo() {
  this.postTimer = 0;
  this.Interval = 1000; // 1 seconds
  this.flyerIDs = {};
};

flyerInfo.prototype = {
  stopGetTimer : function() {
    clearTimeout(this.postTimer);
    this.postTimer=0;
  },
  startGet: function(){
    this.stopGetTimer(); // avoid duplicate timer
    this.postTimer=setTimeout(this.get.bind(this), this.Interval);
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
      if ( !this.flyerIDs[flyer.properties.id] 
        && flyer.properties.startAt <= now
        && now <= flyer.properties.endAt
      ){
        //console.log("==>WRITE CIRCLE!")
        this.flyerIDs[flyer.properties.id]=true;
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
  get          : function() {
    let bounds=map.getBounds();
    //console.log(bounds);
    let url='/api/flyers?south='+bounds.getSouthWest().lat()+'&west='+bounds.getSouthWest().lng()+'&north='+bounds.getNorthEast().lat()+'&east='+bounds.getNorthEast().lng();
    doHttp('GET',url,this.request,this.drawResponse.bind(this));
    this.postTimer=setTimeout(this.get.bind(this), this.Interval);
  },
  post         : function(id,validPeriod,lat,lng,title,distance){
    flyer={
      "storeId"     : id,
      "title"       : title,
      "validPeriod" : validPeriod,
      "latitude"    : lat,
      "longitude"   : lng,
      "distance"    : distance,
      "stocked"     : 10,
    };
    let url='/api/flyers';
    doHttp('POST',url,flyer,function(){});
  }
}

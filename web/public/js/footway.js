var displayFootway = false;

var toggleFootway = function() {
  if (! displayFootway){
    drawFootway();
    displayFootway = true;
  }else{
    removeFootway();
    displayFootway = false;
  }
};

var isFootway = function(tags) {
  if (tags.building ) return false;
  return ( tags.highway == "footway"
    || tags.highway == "pedestrian"
    || tags.highway == "steps"
    || tags.highway == "path"
    || tags.highway == "unclassified"
    || tags.highway == "primary"
    || tags.highway == "trunk"
    || tags.highway == "trunk_link"
    || tags.highway == "tertiary"
    || tags.highway == "service"
    || tags.highway == "residential"
    || tags.sidewalk == "both"
    || tags.sidewalk == "left"
    || tags.sidewalk == "right"
    || tags.foot == "yes"
    || tags.indoor == "yes"
    || tags.bridge == "viaduct"
    || tags.public_transport == "platform"
    || tags.railway == "platform");
};

var drawFootway = function() {
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
            radius: 1
          });
          addFootway(circle)

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
        strokeOpacity: 0.8,
        strokeWeight: 2
        });
      addFootway(polyline)

      polyline.addListener('click', function(e) {
        console.log("Way:===============");
        way.strokeWeight=5
        for(prop in way.Tags){
          console.log(" "+prop+": "+way.Tags[prop]);
        }
      });
    }
  }
  doPost('/footway',null,callback);
};

var footways=[];

var addFootway = function(footway){
  footway.setMap(map);
  var now = Date.now()
  footways.push({time:now, shape:footway});
};

var removeFootway = function(){
  while( footways[0] ){
    let v = footways[0]
    v.shape.setMap(null)
    footways.shift()
  }
};


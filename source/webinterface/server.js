var express = require('express');
var bodyParser = require('body-parser'); 
var app = express();
var jsonfile = require('jsonfile');

app.use(express.static('public'));

var alarm1time = JSON.parse(JSON.stringify(jsonfile.readFileSync('./public/json/alarm1.json')))["time"];
var alarm2time = JSON.parse(JSON.stringify(jsonfile.readFileSync('./public/json/alarm2.json')))["time"];
var alarm3time = JSON.parse(JSON.stringify(jsonfile.readFileSync('./public/json/alarm3.json')))["time"];
var alarm4time = JSON.parse(JSON.stringify(jsonfile.readFileSync('./public/json/alarm4.json')))["time"];
app.use(bodyParser.urlencoded({
  extended: true
}));

app.get('/', function(req, res) {
    res.sendFile(path.join(__dirname + '/index.html'));
});


app.post(
 
  // Route 
  '/',
   // Express request-handler now receives filtered and validated data 
   function(req, res){
   	if(req.body.mytime1== "") {
   		var mytime1 = alarm1time;
   	} else {
   		var mytime1 = req.body.mytime1;
   	}
   	if(req.body.sound1 == "on") {
   		var sound1 = "on";
   	} else {
   		var sound1 = "off";
   	}
   	if(req.body.vibration1 == "on") {
   		var vibration1 = "on";
   	} else {
   		var vibration1 = "off";
   	}
   	if(req.body.mytime2== "") {
   		var mytime2 = alarm2time;
   	} else {
   		var mytime2 = req.body.mytime2;
   	}
   	if(req.body.sound2 == "on") {
   		var sound2 = "on";
   	} else {
   		var sound2 = "off";
   	}
   	if(req.body.vibration2 == "on") {
   		var vibration2 = "on";
   	} else {
   		var vibration2 = "off";
   	}
	if(req.body.mytime3== "") {
		var mytime3 = alarm3time;
	} else {
		var mytime3 = req.body.mytime3;
	}
	if(req.body.sound3 == "on") {
		var sound3 = "on";
	} else {
		var sound3 = "off";
	}
	if(req.body.vibration3 == "on") {
		var vibration3 = "on";
	} else {
		var vibration3 = "off";
	}
      if(req.body.mytime4== "") {
      	var mytime4 = alarm4time;
      } else {
      	var mytime4 = req.body.mytime4;
      }

      if(req.body.sound4 == "on") {
      	var sound4 = "on";
      } else {
      	var sound4 = "off";
      }
      if(req.body.vibration4 == "on") {
      	var vibration4 = "on";
      } else {
      	var vibration4 = "off";
      }
	var newalarm1 = {"id":"alarm1", "time":mytime1, "sound":sound1, "vibration":vibration1};
	var newalarm2 = {"id":"alarm2", "time":mytime2, "sound":sound2, "vibration":vibration2};
	var newalarm3 = {"id":"alarm3", "time":mytime3, "sound":sound3, "vibration":vibration3};
	var newalarm4 = {"id":"alarm4", "time":mytime4, "sound":sound4, "vibration":vibration4};
	jsonfile.writeFile('./public/json/alarm1.json', newalarm1); 
	jsonfile.writeFile('./public/json/alarm2.json', newalarm2); 
	jsonfile.writeFile('./public/json/alarm3.json', newalarm3); 
	jsonfile.writeFile('./public/json/alarm4.json', newalarm4); 
	res.redirect('back');
  }
);
app.listen(3000);
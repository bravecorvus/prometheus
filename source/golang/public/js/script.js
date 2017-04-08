$.getJSON("/json/alarms.json", function( data ) {
    var time;
    var sound;
    var vibration;
    var soundid = "#alarmsound";
    var vibid = "#alarmvibration";
    var currentint;
    data[0].time;
    data[0].sound;
    data[0].vibration;
    for (let i = 0; i < 4; i++) {
        currentint = i + 1;
        time = data[i].time;
        sound = data[i].sound;
        vibration = data[i].vibration;
        document.getElementById("tit".concat(currentint)).innerHTML = "<h1>"+time+"</h1>";
        if(sound == "on") {
            $(soundid.concat(currentint)).append('<input id="sound'+currentint+'" name="sound'+currentint+'" type="checkbox" value="on" checked><div class="slider round" onClick = "document.getElementById("'.concat('alarm'.concat(currentint.concat('sound").submit();"></div>'))))
        } else {
            $(soundid.concat(currentint)).append('<input id="sound'+currentint+'" name="sound'+currentint+'" type="checkbox" value="off" onClick = "document.getElementById("alarm'+currentint+'sound").submit();"><div class="slider round"></div>')
        }
        if(vibration == "on") {
            $(vibid.concat(currentint)).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="on" checked><div class="slider round" onClick = "document.getElementById("alarm'+currentint+'vibration").submit();"></div>')

        } else {
            $(vibid.concat(currentint)).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="off" onClick = "document.getElementById("alarm'+currentint+'vibration").submit();"><div class="slider round"></div>')
        }
    }
});
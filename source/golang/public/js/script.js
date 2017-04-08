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
            $(soundid.concat(currentint)).append('<input id="sound'+currentint+'" name="sound'+currentint+'" type="checkbox" value="on" checked><div class="slider round"></div>')
        } else {
            $(soundid.concat(currentint)).append('<input id="sound'+currentint+'" name="sound'+currentint+'+" type="checkbox" value="off"><div class="slider round"></div>')
        }
        if(vibration == "on") {
            $(vibid.concat(currentint)).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="on" checked><div class="slider round"></div>')

        } else {
            $(vibid.concat(currentint)).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="off"><div class="slider round"></div>')
        }
    }
});

$(document).ready(function() {
    document.getElementById("sound1").setAttribute("onClick", "document.getElementById('alarm1sound').submit();");
    document.getElementById("vibration1").setAttribute("onClick", "document.getElementById('alarm1vibration').submit();");
    document.getElementById("sound2").setAttribute("onClick", "document.getElementById('alarm2sound').submit();");
    document.getElementById("vibration2").setAttribute("onClick", "document.getElementById('alarm2vibration').submit();");
    document.getElementById("sound3").setAttribute("onClick", "document.getElementById('alarm3sound').submit();");
    document.getElementById("vibration3").setAttribute("onClick", "document.getElementById('alarm3vibration').submit();");
    document.getElementById("sound4").setAttribute("onClick", "document.getElementById('alarm4sound').submit();");
    document.getElementById("vibration4").setAttribute("onClick", "document.getElementById('alarm4vibration').submit();");
});
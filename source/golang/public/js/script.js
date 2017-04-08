function submitalarm1sound() {
    document.getElementById('alarm1sound').submit();
}

function submitalarm1vibration() {
    document.getElementById('alarm1vibration').submit();
}

function submitalarm2sound() {
    document.getElementById('alarm2sound').submit();
}

function submitalarm2vibration() {
    document.getElementById('alarm2vibration').submit();
}

function submitalarm3sound() {
    document.getElementById('alarm3sound').submit();
}

function submitalarm3vibration() {
    document.getElementById('alarm3vibration').submit();
}

function submitalarm4sound() {
    document.getElementById('alarm4sound').submit();
}

function submitalarm4vibration() {
    document.getElementById('alarm4vibration').submit();
}


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
        document.getElementById("tit1").innerHTML = "<h1>"+time+"</h1>";
        if(sound == "on") {
            $(soundid+currentint).append('<input id="sound'+currentint+'" name="sound'+currentint+'" type="checkbox" value="on" checked><div class="slider round"></div>')
        } else {
            $(soundid+currentint).append('<input id="sound'+currentint+'" name="sound'+currentint'+" type="checkbox" value="off"><div class="slider round"></div>')
        }
        if(vibration == "on") {
            $(vibid+currentint).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="on" checked><div class="slider round"></div>')

        } else {
            $(vibid+currentint).append('<input id="vibration'+currentint+'" name="vibration'+currentint+'" type="checkbox" value="off"><div class="slider round"></div>')
        }
    }
});

$(document).ready(function() {
    document.getElementById("sound1").setAttribute("onClick", "submitalarm1sound()");
    document.getElementById("vibration1").setAttribute("onClick", "submitalarm1vibration()");
    document.getElementById("sound2").setAttribute("onClick", "submitalarm2sound()");
    document.getElementById("vibration2").setAttribute("onClick", "submitalarm2vibration()");
    document.getElementById("sound3").setAttribute("onClick", "submitalarm3sound()");
    document.getElementById("vibration3").setAttribute("onClick", "submitalarm3vibration()");
    document.getElementById("sound4").setAttribute("onClick", "submitalarm4sound()");
    document.getElementById("vibration4").setAttribute("onClick", "submitalarm4vibration()");
});
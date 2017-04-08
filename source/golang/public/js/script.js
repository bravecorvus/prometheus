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




// $.getJSON( "/json/alarm1.json", function( data ) {
//     var alarm1time = data[0].time;
//     var alarm1sound = data[0].sound;
//     var alarm1vibration = data[0].vibration;
//     document.getElementById("tit1").innerHTML = "<h1>"+alarm1time+"</h1>";
//     if(alarm1sound == "on") {
//         $("#alarmsound1").append('<input id="sound1" name="sound1" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmsound1").append('<input id="sound1" name="sound1" type="checkbox" value="off"><div class="slider round"></div>')
//     }
//     if(alarm1vibration == "on") {
//         $("#alarmvibration1").append('<input id="vibration1" name="vibration1" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmvibration1").append('<input id="vibration1" name="vibration1" type="checkbox" value="off"><div class="slider round"></div>')
//     }
// });


// $.getJSON( "/json/alarm2.json", function( data ) {
//     var alarm2time = data[0].time;
//     var alarm2sound = data[0].sound;
//     var alarm2vibration = data[0].vibration;
//     document.getElementById("tit2").innerHTML = "<h1>"+alarm2time+"</h1>";
//     if(alarm2sound == "on") {
//         $("#alarmsound2").append('<input id="sound2" name="sound2" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmsound2").append('<input id="sound2" name="sound2" type="checkbox" value="off"><div class="slider round"></div>')
//     }
//     if(alarm2vibration == "on") {
//         $("#alarmvibration2").append('<input id="vibration2" name="vibration2" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmvibration2").append('<input id="vibration2" name="vibration2" type="checkbox" value="off"><div class="slider round"></div>')
//     }
// });

// $.getJSON( "/json/alarm3.json", function( data ) {
//     var alarm3time = data[0].time;
//     var alarm3sound = data[0].sound;
//     var alarm3vibration = data[0].vibration;
//     document.getElementById("tit3").innerHTML = "<h1>"+alarm3time+"</h1>";
//     if(alarm3sound == "on") {
//         $("#alarmsound3").append('<input id="sound3" name="sound3" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmsound3").append('<input id="sound3" name="sound3" type="checkbox" value="off"><div class="slider round"></div>')
//     }
//     if(alarm3vibration == "on") {
//         $("#alarmvibration3").append('<input id="vibration3" name="vibration3" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmvibration3").append('<input id="vibration3" name="vibration3" type="checkbox" value="off"><div class="slider round"></div>')
//     }
// });

// $.getJSON( "/json/alarm4.json", function( data ) {
//     var alarm4time = data[0].time;
//     var alarm4sound = data[0].sound;
//     var alarm4vibration = data[0].vibration;
//     document.getElementById("tit4").innerHTML = "<h1>"+alarm4time+"</h1>";
//     if(alarm4sound == "on") {
//         $("#alarmsound4").append('<input id="sound4" name="sound4" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmsound4").append('<input id="sound4" name="sound4" type="checkbox" value="off"><div class="slider round"></div>')
//     }
//     if(alarm4vibration == "on") {
//         $("#alarmvibration4").append('<input id="vibration4" name="vibration4" type="checkbox" value="on" checked><div class="slider round"></div>')

//     } else {
//         $("#alarmvibration4").append('<input id="vibration4" name="vibration4" type="checkbox" value="off"><div class="slider round"></div>')
//     }
// });

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
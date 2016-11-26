$.getJSON( "alarm1.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm1time = items[1];
    var alarm1sound = items[2];
    var alarm1vibration = items[3];
    document.getElementById("tit1").innerHTML = "<h1>"+alarm1time+"</h1>";
    if(alarm1sound == "on") {
        sound1toggleOn();
    } else {
        sound1toggleOff();
    }
    if(alarm1vibration == "on") {
        vibration1toggleOn();
    } else {
        vibration1toggleOff();
    }
});


$.getJSON( "alarm2.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm2time = items[1];
    var alarm2sound = items[2];
    var alarm2vibration = items[3];
    document.getElementById("tit2").innerHTML = "<h1>"+alarm2time+"</h1>";
    if(alarm2sound == "on") {
        sound2toggleOn();
    } else {
        sound2toggleOff();
    }
    if(alarm2vibration == "on") {
        vibration2toggleOn();
    } else {
        vibration2toggleOff();
    }
});

$.getJSON( "alarm3.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm3time = items[1];
    var alarm3sound = items[2];
    var alarm3vibration = items[3];
    document.getElementById("tit3").innerHTML = "<h1>"+alarm3time+"</h1>";
    if(alarm3sound == "on") {
        sound3toggleOn();
    } else {
        sound3toggleOff();
    }
    if(alarm3vibration == "on") {
        vibration3toggleOn();
    } else {
        vibration3toggleOff();
    }
});

$.getJSON( "alarm4.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm4time = items[1];
    var alarm4sound = items[2];
    var alarm4vibration = items[3];
    document.getElementById("tit4").innerHTML = "<h1>"+alarm4time+"</h1>";
    if(alarm4sound == "on") {
        sound4toggleOn();
    } else {
        sound4toggleOff();
    }
    if(alarm4vibration == "on") {
        vibration4toggleOn();
    } else {
        vibration4toggleOff();
    }
});


function update_alarm1() {
        var t = document.getElementById('mytime1').value;
        if($('#sound1:checked').val()==null) {
            var s = "off";
        } else {
            var s = $('#sound1:checked').val();
        }
        if($('#vibration1:checked').val()==null) {
            var v = "off";
        } else {
            var v = $('#vibration1:checked').val();
        }
        if(t != ""){
            document.getElementById('tit1').innerHTML = "<h1>"+t+"</h1>";
        }
    }
function update_alarm2() {
        var t = document.getElementById('mytime2').value;
        if($('#sound2:checked').val()==null) {
            var s = "off";
        } else {
            var s = $('#sound2:checked').val();
        }
        if($('#vibration2:checked').val()==null) {
            var v = "off";
        } else {
            var v = $('#vibration2:checked').val();
        }
        if(t != ""){
            document.getElementById('tit2').innerHTML = "<h1>"+t+"</h1>";
        }
    }

function update_alarm3() {
        var t = document.getElementById('mytime3').value;
        if($('#sound3:checked').val()==null) {
            var s = "off";
        } else {
            var s = $('#sound3:checked').val();
        }
        if($('#vibration3:checked').val()==null) {
            var v = "off";
        } else {
            var v = $('#vibration3:checked').val();
        }
        if(t != ""){
            document.getElementById('tit3').innerHTML = "<h1>"+t+"</h1>";
        }
    }

function update_alarm4() {
        var t = document.getElementById('mytime4').value;
        // var s = document.getElementById('sound4').value;
        if($('#sound4:checked').val()==null) {
            var s = "off";
        } else {
            var s = $('#sound4:checked').val();
        }
        if($('#vibration4:checked').val()==null) {
            var v = "off";
        } else {
            var v = $('#vibration4:checked').val();
        }
        if(t != ""){
            document.getElementById('tit4').innerHTML = "<h1>"+t+"</h1>";
        }
}
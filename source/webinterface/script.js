$.getJSON( "alarm1.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm1time = items[1];
    var alarm1sound = items[2];
    var alarm1vibration = items[3];
    document.getElementById("tit1").innerHTML = "<h1>"+alarm1time+"</h1>";
    if(alarm1sound != "on") {
        document.getElementById('sound1').click();
    }
    if(alarm1vibration != "on") {
        document.getElementById('vibration1').click();
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
    if(alarm2sound != "on") {
        document.getElementById('sound2').click();
    } 
    if(alarm2vibration != "on") {
        document.getElementById('vibration2').click();
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
    if(alarm3sound != "on") {
        document.getElementById('sound3').click();
    } 
    if(alarm3vibration != "on") {
        document.getElementById('vibration3').click();
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
    if(alarm4sound != "on") {
        document.getElementById('sound4').click();
    } 
    if(alarm4vibration != "on") {
        document.getElementById('vibration4').click();
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
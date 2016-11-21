$.getJSON( "alarm1.json", function( data ) {
    var items = [];
  $.each( data, function( key, val ) {
    items.push(val);
  });
    var alarm1time = items[1];
    var alarm1sound = items[2];
    var alarm1vibration = items[3];
    document.getElementById("tit1").innerHTML = "<h1>"+alarm1time+"</h1>";
    console.log(alarm1sound);
    console.log(alarm1vibration);
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
        console.log(t);
        console.log(s);
        console.log(v);
        document.getElementById('tit1').innerHTML = "<h1>"+t+"</h1>";
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
        console.log(t);
        console.log(s);
        console.log(v);
        document.getElementById('tit2').innerHTML = "<h1>"+t+"</h1>";
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
        console.log(t);
        console.log(s);
        console.log(v);
        document.getElementById('tit3').innerHTML = "<h1>"+t+"</h1>";
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
        console.log(t);
        console.log(s);
        console.log(v);
        document.getElementById('tit4').innerHTML = "<h1>"+t+"</h1>";
        // var textToSave = '{"id":"alarm4", "time":"'+t+'", "sound":"' +s+ '", "vibration":"' +v+ '"}';
}
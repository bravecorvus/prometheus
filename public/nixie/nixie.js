
/****************************************************************************

    Virtual nixie tube display, clock & calculator DHTML components

    v 1.05, 20080214a

    (c) 2007-08 Cestmir Hybl, cestmir.hybl@nustep.net
    http://cestmir.freeside.sk/projects/dhtml-nixie-display

    license: free for non-commercial use, copyright must be preserved

 ****************************************************************************/



/*   NixieDisplay   */

// public class NixieDisplay
function NixieDisplay()
{
  // public
  this.id = 'nixie';
  this.elContainer = null;
  this.charCount = 10;
  this.autoDecimalPoint = true;  // automatically extracts decimal point index in setText() call
  this.align = 'left';           // alignment of text via setText() call
  this.afterUpdate = null;   // after display update callback

  this.charWidth = 62;
  this.charHeight = 150;
  this.charGapWidth = 0;
  this.extraGapsWidths = [];
  this.createCharElements = true;

  this.text = '';
  this.decimalPoint = -1;

  this.urlCharsetImage = 'nixie/zm1082_l1_09bdm_62x150_8b.png';
  this.charMap = {
      0: 0,   1: 1,   2: 2,   3: 3,   4: 4,   5: 5,   6: 6,   7: 7,   8: 8,   9: 9,
    '0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9, ' ': 10, '-': 11,
    'default': 10
  };
    // maps displayable chars onto glyph matrix indexes

  // protected
  function _drawChar(index)
  {
    var el = document.getElementById(this.id + '_d' + index);
    var charIndex = this.charMap[this.text.charAt(index)];
    if (!charIndex && charIndex !== 0)
      charIndex = this.charMap['default'];
    var x = - (charIndex * this.charWidth);
    var y = (index === this.decimalPoint ? - this.charHeight : 0);
    el.style.backgroundPosition = x + 'px ' + y + 'px';
  }
  this._drawChar = _drawChar;

  // Shows given string on display
  // public
  function setText(text, updateDecimalPoint)
  {
    // force string type
    this.text = text + '';

    // extract decimal point
    updateDecimalPoint = (typeof(updateDecimalPoint) != 'undefined' ? updateDecimalPoint : this.autoDecimalPoint);
    if (updateDecimalPoint) {
      var i = this.text.indexOf('.');
      if (i >= 0) {
        this.decimalPoint = i - 1;
        // alert(this.decimalPoint);
        this.text = this.text.substr(0, i) + this.text.substr(i + 1);
      } else
        this.decimalPoint = -1;
    }

    // pad up to display width (from left/right acording to this.align)
    if (this.text.length < this.charCount) {
      var pad = '';
      var padWidth = this.charCount - this.text.length;
      for (var i = 0; i < padWidth; i++)
        pad += ' ';
      if (this.align == 'left')
        this.text = this.text + pad;
      else {
        if (this.decimalPoint >= 0)
          this.decimalPoint += padWidth;
        this.text = pad + this.text;
      }
    }

    if (this.text.length > this.charCount)
      this.text = this.text.substr(0, this.charCount);

    // draw chars
    for (var i = 0; i < this.text.length; i++) {
      this._drawChar(i);
    }

    if (this.afterUpdate)
      this.afterUpdate(this);
  }
  this.setText = setText;

  // Sets char at given display position
  // public
  function setChar(index, chr)
  {
    // alert(chr);
    this.text = this.text.substring(0, index) + chr + this.text.substring(index + 1);
    this.setText(this.text, false);
  }
  this.setChar = setChar;

  function setDecimalPoint(index)
  {
    var oldDecimalPoint = this.decimalPoint;
    this.decimalPoint = ((!index && index !== 0) ? -1 : index);
    if (oldDecimalPoint != this.decimalPoint) {
      if (oldDecimalPoint >= 0)
        this._drawChar(oldDecimalPoint);
      if (this.decimalPoint >= 0)
        this._drawChar(this.decimalPoint);
    }
  }
  this.setDecimalPoint = setDecimalPoint;

  // Clears display - fills all positions with given char (space by default).
  // public
  function clear(chr)
  {
    chr = (typeof(chr) == 'undefined' ? ' ' : chr);
    this.text = '';
    for (var i = 0; i < this.charCount; i++)
      this.text += chr;
    this.decimalPoint = -1;
    this.setText(this.text);
  }
  this.clear = clear;

  // Shifts display contents left or right
  // public
  function shift(direction, step)
  {
    step = (!step && step !== 0 ? 1 : step);
    direction = (!direction ? 'left' : direction);

    if (this.decimalPoint >= 0) {
      this.decimalPoint += (direction == 'left' ? - step : + step);
      if (this.decimalPoint >= this.charCount)
        this.decimalPoint = -1;
    }

    if (direction == 'left')
      this.text = this.text.substr(step) + ' '; // @todo padding for step != +/-1
    else if (direction == 'right')
      this.text = ' ' + this.text.substr(0, this.text.length - 1); // @todo padding for step != +/-1
    this.setText(this.text, false);
  }
  this.shift = shift;

  // public
  function init()
  {
    if (!this.elContainer) {
      this.elContainer = document.getElementById(this.id);
      if (!this.elContainer)
        throw "Container element '" + this.id + "' not found";
    }
    this.elContainer.style.position = 'relative';

    if (this.createCharElements) {
      var totalWidth = 0;
      for (var i = 0; i < this.charCount; i++) {
        var charWidthIncludingGap = (this.charWidth + this.charGapWidth);

        var elId = this.id + '_d' + i;
        var el0 = document.getElementById(elId);
        var el = (el0 ? el0 : document.createElement('div'));
        el.id = this.id + '_d' + i;
        el.className = 'digit d' + i;
        el.style.position = 'absolute';
        el.style.left = totalWidth + 'px';
        el.style.width = this.charWidth + 'px';
        el.style.height = this.charHeight + 'px';
        el.style.background = 'url(' + this.urlCharsetImage + ')';
        if (!el.parentNode)
          this.elContainer.appendChild(el);

        totalWidth += charWidthIncludingGap + (this.extraGapsWidths[i] ? this.extraGapsWidths[i] : 0);
      }
      this.elContainer.style.width = totalWidth + 'px';
      this.elContainer.style.height = this.charHeight + 'px';
    }

    if (this.text)
      this.setText(this.text)
    else
      this.clear();
  }
  this.init = init;
}



/*   NixieClock   */

// public class NixieClock : NixieDisplay
function NixieClock()
{
  // public

  // private
  this.lastSeconds = -1;

  // Show current time on "display"
  // public
  function showCurrentTime(refreshAfterChangeOnly)
  {
    var d = new Date();

    var s = d.getSeconds();

    if (refreshAfterChangeOnly && s == this.lastSeconds)
      return;
    else
      this.lastSeconds = s;

    var h = d.getHours();
    if (h == 0) {
      h = 12;
    }
    else if (h == 13) {
      h = 1;
    }
    else if (h == 14) {
      h = 2;
    }
    else if (h == 15) {
      h = 3;
    }
    else if (h == 16) {
      h = 4;
    }
    else if (h == 17) {
      h = 5;
    }
    else if (h == 18) {
      h = 6;
    }
    else if (h == 19) {
      h = 7;
    }
    else if (h == 20) {
      h = 8;
    }
    else if (h == 21) {
      h = 9;
    }
    else if (h == 22) {
      h = 10;
    }
    else if (h == 23) {
      h = 11;
    }
    var m = d.getMinutes();

    var digits = '';

    digits += (h / 10) | 0;
    digits += h % 10;
    digits += (m / 10) | 0;
    digits += m % 10;
    digits += (s / 10) | 0;
    digits += s % 10;

    this.setText(digits);
  }
  this.showCurrentTime = showCurrentTime;

  // Run clock (via scheduling a periodic callback to showCurrentTime())
  // public
  function run()
  {
    if (!this.elContainer)
      this.init();
    var __nixieClock = this;
    window.setInterval(function() { __nixieClock.showCurrentTime(true); }, 100);
  }
  this.run = run;

  this.ancestor = NixieDisplay;
  this.ancestor();

  this.charCount = 6;
  this.extraGapsWidths[1] = 20;
  this.extraGapsWidths[3] = 20;
}



/*  NixieCalculator  */

// @todo rounding of rightmost digit

// public class NixieCalculator : NixieDisplay
function NixieCalculator()
{
  // public
  this.id = 'nixieCalc';
  this.digitCount = 13;
  this.display = new NixieDisplay();

  // private
  this.operandStack = [];
  this.newValueAtNextChar = false;
  this.fullPrecisionValue = 0;

  // private
  function push(value)
  {
    this.operandStack[this.operandStack.length] = value; // JS50 compatible .push()
  }
  this.push = push;

  // private
  function pop()
  {
    if (!this.operandStack.length)
      return null;
    var v = this.operandStack[this.operandStack.length - 1];
    this.operandStack = this.operandStack.slice(0, this.operandStack.length - 1); // JS50 compatible .pop()
    return v;
  }
  this.pop = pop;

  // public
  function getValue()
  {
    if (this.fullPrecisionValue !== null)
      return this.fullPrecisionValue;

    var v = this.display.text;

    // insert decimal point
    if (this.display.decimalPoint >= 0 && this.display.decimalPoint < this.digitCount - 1)
      v = v.substr(0, this.display.decimalPoint + 1) + '.' + v.substr(this.display.decimalPoint + 1);

    // remove padding spaces
    var i = 0;
    while (i < v.length && v.charAt(i) == ' ')
      i++;
    v = v.substr(i);

    // convert to number
    v = parseFloat(v);

    return v;
  }
  this.getValue = getValue;

  // public
  function setValue(v)
  {
    if (typeof(v) != 'number')
      v = parseFloat(v);
    if (isNaN(v) || v > this.maxNumber || v < -this.maxNumber)
      this.error();
    else {
      this.fullPrecisionValue = v;
      if (v.toFixed) {
        // force fixed-point notation (JS5.5+)
        var s = (v >= 0 ? ' ' : '') + v.toFixed(1);
        s = s.substring(0, s.length - 2);
        // (s now contains string with integer part of value, prefixed by either ' ' or '-')
        v = v.toFixed(this.digitCount - s.length); // to fixed point + round rightmost digit
      } else {
        v = v.toString();
        if (v.toLowerCase().indexOf('e') >= 0) {
          // we won't handle exp notation in JS<5.5
          this.error();
          return;
        }
      }
      if (v !== '0') {
        if (v.charAt(0) != '-')
          v = ' ' + v;
        var c = this.digitCount + (v.indexOf('.') >= 0 ? 1 : 0);
        if (v.length > c)
          v = v.substr(0, c);
        v = v.replace(/^(.{1,}?)\.?0+$/g, '$1'); // strip zero's from right
      }
      this.display.setText(v);
    }
  }
  this.setValue = setValue;

  // private
  function eval(v1, o, v2)
  {
    try {
      switch (o) {
        case '+':
          return v1 + v2;
        case '-':
          return v1 - v2;
        case '*':
          return v1 * v2;
        case '/':
          return v1 / v2;
        case '^':
          return Math.pow(v1, v2);
        case 'sqrt':
          return Math.sqrt(v1);
        case 'sqr':
          return v1 * v1;
        default:
          throw "Unsupported operand: '" + o + "'";
      }
    } catch(e) {
      this.error();
    }
  }
  this.eval = eval;

  // public
  function error()
  {
    var s= '';
    for (var i = 0; i < this.digitCount; i++)
      s += '-';
    this.operandStack = [];
    this.newValueAtNextChar = true;
    this.fullPrecisionValue = null;
    this.display.setText(s);
  }
  this.error = error;

  // public
  function clear()
  {
    this.display.clear();
    this.setValue(0);
    this.operandStack = [];
    this.fullPrecisionValue = null;
  }
  this.clear = clear;

  // public
  function keyDown(event0)
  {
    var e = (event0 ? event0 : event);
    var k = e.keyCode;

    var cancelEvent = true;

    if (k == 8) {
      // backspace
      if (this.display.text.charAt(this.digitCount - 2) == ' ')
        this.display.setChar(this.digitCount - 1, '0');
      else
        this.display.shift('right');
      this.fullPrecisionValue = null;
    } else if (k == 27) {
      // escape
      this.clear();
    } else
      cancelEvent = false;

    return !cancelEvent;
  }
  this.keyDown = keyDown;

  // public
  function keyPress(event0)
  {
    var e = (event0 ? event0 : event);
    var k = (e.keyCode ? e.keyCode : e.which); // IE: .keyCode, FF: .which
    var chr = String.fromCharCode(k);

    var cancelEvent = true;

    var newValueAtThisChar =  this.newValueAtNextChar;
    this.newValueAtNextChar = true;

    if (chr >= '0' && chr <= '9') {
      this.fullPrecisionValue = null;
      if (newValueAtThisChar) {
        this.display.clear();
      }
      if (this.display.text.charAt(1) == ' ' || this.display.text.charAt(1) == '-') {
        if (this.display.text.charAt(this.digitCount - 1) == '0' && this.display.text.charAt(this.digitCount - 2) == ' ' && this.display.decimalPoint < 0)
          ;
        else
          this.display.shift('left');
        this.display.setChar(this.digitCount - 1, chr);
      }
      this.newValueAtNextChar = false;
    }
    else if (chr == '.' || chr == ',') {
      this.fullPrecisionValue = null;
      if (newValueAtThisChar)
        this.display.setText(0);
      if (this.display.decimalPoint < 0)
        this.display.setDecimalPoint(this.digitCount - 1);
      this.newValueAtNextChar = false;
    }
    else if (chr == '+' || chr == '-' || chr == '*' || chr == '/' || chr == '^') {
      if (this.operandStack.length > 2)
        // cancel repeated evaluation
        this.operandStack = [];
      if (this.operandStack.length == 2) {
        // previous expression without explicit '=', evaluate it
        this.setValue(this.eval(this.operandStack[this.operandStack.length - 2], this.operandStack[this.operandStack.length - 1], this.getValue()));
        this.operandStack = [];
      }
      // push left operand
      this.push(this.getValue());
      // push operator
      this.push(chr);
    }
    else if (chr == 'm' || chr == 'M') {
      this.setValue(- this.getValue());
      this.newValueAtNextChar = false;
    }
    else if (chr == 'p' || chr == 'P') {
      this.setValue(Math.PI);
    }
    else if (chr == 'q') {
      this.setValue(this.eval(this.getValue(), 'sqrt', null));
    }
    else if (chr == 'Q') {
      this.setValue(this.eval(this.getValue(), 'sqr', null));
    }
    else if (k == 13 || chr == '=') {
      if (this.operandStack.length >= 2) {
        if (this.operandStack.length <= 2)
          // push right operand
          this.push(this.getValue())
        else
          // repeated evaluation (e.g. [1] [+] [1] [=] [=] ...), replace left operand with current display value
          this.operandStack[this.operandStack.length - 3] = this.getValue();
        // alert(this.operandStack);
        var result = this.eval(this.operandStack[this.operandStack.length - 3], this.operandStack[this.operandStack.length - 2], this.operandStack[this.operandStack.length - 1]);
        this.setValue(result);
        // this.operandStack = [];
      }
    }
    // else if (k == 27) {  // Fix: chromium: Esc not handled in keyPress(), http://code.google.com/p/chromium/issues/detail?id=12744
    //   // escape
    //   this.clear();
    // }
    else {
      cancelEvent = false;
      this.newValueAtNextChar = false;
    }

    return !cancelEvent;
  }
  this.keyPress = keyPress;

  // public
  function init()
  {
    this.display.id = this.id;
    this.display.charCount = this.digitCount;
    this.display.align = 'right';
    this.display.init();
    this.display.setText('0');
    this.maxNumber = Math.pow(10, this.digitCount - 1) - 1;
  }
  this.init = init;

  // assign default glyph matrix
  this.display.urlCharsetImage = 'nixie/zm1080_l2_09bdm_45x75_8b.png';
  this.display.charWidth = 45;
  this.display.charHeight = 75;
  this.display.charGapWidth = 5;
}

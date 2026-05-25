// TS port of the 2008-era DHTML nixie tube clock library by Cestmir Hybl
// (cestmir.freeside.sk/projects/dhtml-nixie-display). Only the clock face is
// kept (the original ships a calculator we don't need). It paints six digit
// <div>s into a container, each backed by a sprite-sheet showing one of the
// nixie tubes (0-9 + blank). showCurrentTime() repositions each sprite's
// background-position to the right digit. Cheap, smooth, dependency-free.

const CHAR_MAP: Record<string, number> = {
  "0": 0, "1": 1, "2": 2, "3": 3, "4": 4,
  "5": 5, "6": 6, "7": 7, "8": 8, "9": 9,
  " ": 10, "-": 11,
};

interface NixieClockOptions {
  container: HTMLElement;
  spriteUrl: string;
  charWidth: number;
  charHeight: number;
  charGapWidth?: number;
  extraGapsWidths?: number[];
}

export class NixieClock {
  private container: HTMLElement;
  private spriteUrl: string;
  private charWidth: number;
  private charHeight: number;
  private charGapWidth: number;
  private extraGapsWidths: number[];
  private digits: HTMLDivElement[] = [];
  private lastSeconds = -1;
  private intervalId: number | null = null;
  private text = "";

  constructor(opts: NixieClockOptions) {
    this.container = opts.container;
    this.spriteUrl = opts.spriteUrl;
    this.charWidth = opts.charWidth;
    this.charHeight = opts.charHeight;
    this.charGapWidth = opts.charGapWidth ?? 0;
    this.extraGapsWidths = opts.extraGapsWidths ?? [];
  }

  init() {
    this.container.style.position = "relative";
    let totalWidth = 0;
    for (let i = 0; i < 6; i++) {
      const el = document.createElement("div");
      el.style.position = "absolute";
      el.style.left = totalWidth + "px";
      el.style.width = this.charWidth + "px";
      el.style.height = this.charHeight + "px";
      el.style.background = `url(${this.spriteUrl})`;
      this.container.appendChild(el);
      this.digits.push(el);

      totalWidth +=
        this.charWidth +
        this.charGapWidth +
        (this.extraGapsWidths[i] ?? 0);
    }
    this.container.style.width = totalWidth + "px";
    this.container.style.height = this.charHeight + "px";
    this.setText("      ");
  }

  // Native dimensions before any CSS transform scaling.
  size(): { width: number; height: number } {
    let totalWidth = 0;
    for (let i = 0; i < 6; i++) {
      totalWidth +=
        this.charWidth + this.charGapWidth + (this.extraGapsWidths[i] ?? 0);
    }
    return { width: totalWidth, height: this.charHeight };
  }

  private setText(text: string) {
    this.text = text.padEnd(6, " ").slice(0, 6);
    for (let i = 0; i < 6; i++) {
      const charIndex = CHAR_MAP[this.text.charAt(i)] ?? 10;
      const x = -(charIndex * this.charWidth);
      const digit = this.digits[i];
      if (digit) digit.style.backgroundPosition = x + "px 0px";
    }
  }

  private showCurrentTime() {
    const d = new Date();
    const s = d.getSeconds();
    if (s === this.lastSeconds) return;
    this.lastSeconds = s;

    // 12-hour clock (matches the original behavior; the connected NCS314
    // shield is also 12-hour by default).
    let h = d.getHours();
    if (h === 0) h = 12;
    else if (h > 12) h -= 12;
    const m = d.getMinutes();

    const digits =
      ((h / 10) | 0) +
      "" +
      (h % 10) +
      ((m / 10) | 0) +
      (m % 10) +
      ((s / 10) | 0) +
      (s % 10);
    this.setText(digits);
  }

  run() {
    this.showCurrentTime();
    this.intervalId = window.setInterval(() => this.showCurrentTime(), 100);
  }

  stop() {
    if (this.intervalId !== null) {
      window.clearInterval(this.intervalId);
      this.intervalId = null;
    }
    while (this.container.firstChild) {
      this.container.removeChild(this.container.firstChild);
    }
    this.digits = [];
  }
}

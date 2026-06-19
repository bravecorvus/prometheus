import { useEffect, useRef } from "react";
import { NixieClock as NixieClockLib } from "@/lib/nixie";

interface Props {
  onSnooze?: () => void;
}

// React wrapper around the imperative NixieClock class. Mounts once,
// auto-scales to the parent's width via CSS transform, and forwards taps to
// onSnooze (the alarm-stop gesture).
export function NixieClock({ onSnooze }: Props) {
  const stageRef = useRef<HTMLDivElement>(null);
  const innerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const inner = innerRef.current;
    const stage = stageRef.current;
    if (!inner || !stage) return;

    const clock = new NixieClockLib({
      container: inner,
      spriteUrl: "/nixie/zm1080_l2_09bdm_90x150_8b.png",
      charWidth: 90,
      charHeight: 150,
      charGapWidth: 10,
      extraGapsWidths: [0, 12, 0, 12, 0, 0],
    });
    clock.init();
    clock.run();

    const { width: nativeW, height: nativeH } = clock.size();
    // Breathing room around the digits: enough to clear the stage's rounded
    // corners horizontally and the drop-shadow glow vertically. Without these
    // the scaled clock sits flush against the clip region and edges get cut.
    const PAD_X = 16;
    const PAD_Y = 16;
    const fit = () => {
      const availW = stage.clientWidth - PAD_X * 2;
      const scale = Math.min(1, availW / nativeW);
      inner.style.transform = `scale(${scale})`;
      stage.style.height = nativeH * scale + PAD_Y * 2 + "px";
    };
    fit();
    const ro = new ResizeObserver(fit);
    ro.observe(stage);

    return () => {
      ro.disconnect();
      clock.stop();
    };
  }, []);

  return (
    <div
      ref={stageRef}
      className="nixie-stage w-full max-w-3xl mx-auto rounded-2xl"
      onClick={onSnooze}
      role="button"
      aria-label="Tap to snooze"
    >
      <div ref={innerRef} />
    </div>
  );
}

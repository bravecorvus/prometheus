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
    // The nixie lib sets the inner container to position: relative and
    // builds absolutely positioned digit divs inside. We re-anchor it as
    // position: absolute within the stage so we can place + scale it
    // explicitly from the top-left, without relying on flexbox+transform
    // interactions (which clip badly because flex lays out the unscaled box).
    inner.style.position = "absolute";
    inner.style.transformOrigin = "top left";

    // Breathing room around the digits: clears the rounded corners (16px)
    // and gives the orange drop-shadow glow room to bleed.
    const PAD_X = 16;
    const PAD_Y = 16;
    const fit = () => {
      const availW = stage.clientWidth - PAD_X * 2;
      const scale = Math.min(1, availW / nativeW);
      const scaledW = nativeW * scale;
      const scaledH = nativeH * scale;
      inner.style.transform = `scale(${scale})`;
      inner.style.left = (stage.clientWidth - scaledW) / 2 + "px";
      inner.style.top = PAD_Y + "px";
      stage.style.height = scaledH + PAD_Y * 2 + "px";
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

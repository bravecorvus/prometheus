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
    inner.style.position = "absolute";
    inner.style.transformOrigin = "top left";

    // Padding needs to clear the stage's 16px rounded corner radius (digits
    // that fall inside a corner's clip arc get sliced), plus the orange
    // drop-shadow bloom. 24px on each side leaves a safe margin.
    const PAD_X = 24;
    const PAD_Y = 24;
    const fit = () => {
      const stageW = stage.clientWidth;
      if (stageW === 0) return; // pre-layout; ResizeObserver will retry
      const availW = stageW - PAD_X * 2;
      const scale = Math.min(1, availW / nativeW);
      const scaledW = nativeW * scale;
      const scaledH = nativeH * scale;
      inner.style.transform = `scale(${scale})`;
      inner.style.left = (stageW - scaledW) / 2 + "px";
      inner.style.top = PAD_Y + "px";
      stage.style.height = scaledH + PAD_Y * 2 + "px";
    };
    // Initial fit deferred to next frame so stage has its final width after
    // Tailwind utility CSS has settled.
    const raf = requestAnimationFrame(fit);
    const ro = new ResizeObserver(fit);
    ro.observe(stage);

    return () => {
      cancelAnimationFrame(raf);
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

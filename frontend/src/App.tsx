import { useEffect, useState } from "react";
import { type State, type Alarm, getState, snooze } from "@/lib/api";
import { NixieClock } from "@/components/NixieClock";
import { AlarmCard } from "@/components/AlarmCard";
import { SettingsTab } from "@/components/SettingsTab";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { Toaster } from "@/components/ui/sonner";
import { toast } from "sonner";

export default function App() {
  const [state, setState] = useState<State | null>(null);

  useEffect(() => {
    getState().then(setState).catch((err) => {
      toast.error("Failed to load state", { description: String(err) });
    });
  }, []);

  useEffect(() => {
    if (!state?.demo) return;
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(`${proto}//${window.location.host}/ws`);
    ws.onmessage = (e) => {
      let data: { event?: string };
      try {
        data = JSON.parse(e.data);
      } catch {
        return;
      }
      switch (data.event) {
        case "vibration_on":
          toast.warning("Vibration started");
          break;
        case "vibration_off":
          toast.info("Vibration stopped");
          break;
        case "sound_start":
          toast.warning("Sound playing");
          break;
        case "sound_stop":
          toast.info("Sound stopped");
          break;
      }
    };
    return () => ws.close();
  }, [state?.demo]);

  if (!state) {
    return (
      <div className="grid min-h-svh place-items-center text-muted-foreground">
        Loading…
      </div>
    );
  }

  function updateAlarm(name: string, next: Alarm) {
    setState((s) => {
      if (!s) return s;
      return {
        ...s,
        alarms: s.alarms.map((a) => (a.name === name ? next : a)),
      };
    });
  }

  async function onSnooze() {
    try {
      await snooze();
    } catch (err) {
      toast.error("Snooze failed", { description: String(err) });
    }
  }

  return (
    <div className="min-h-svh bg-background">
      <div className="mx-auto max-w-5xl px-4 py-6 space-y-6 sm:px-6 sm:py-8">
        <header className="relative">
          <h1 className="text-2xl font-semibold tracking-tight text-center">
            Prometheus<span className="text-primary">.</span>
          </h1>
          {state.demo && (
            <span className="absolute right-0 top-1/2 -translate-y-1/2 rounded-full bg-secondary px-3 py-1 text-xs uppercase tracking-wider text-secondary-foreground">
              Demo mode
            </span>
          )}
        </header>

        <NixieClock onSnooze={onSnooze} />

        <Tabs defaultValue="home" className="w-full">
          <div className="flex justify-center">
            <TabsList className="grid w-full grid-cols-3 sm:w-auto sm:inline-grid">
              <TabsTrigger value="home">Alarms</TabsTrigger>
              <TabsTrigger value="settings">Settings</TabsTrigger>
              <TabsTrigger value="about">About</TabsTrigger>
            </TabsList>
          </div>

          <TabsContent value="home" className="pt-4">
            <div className="grid gap-4 sm:grid-cols-2">
              {state.alarms.map((alarm, i) => (
                <AlarmCard
                  key={alarm.name}
                  index={i}
                  alarm={alarm}
                  onChange={(next) => updateAlarm(alarm.name, next)}
                />
              ))}
            </div>
          </TabsContent>

          <TabsContent value="settings" className="pt-4">
            <SettingsTab
              state={state}
              onChange={(next) => setState(next)}
            />
          </TabsContent>

          <TabsContent value="about" className="pt-4">
            <div className="overflow-hidden rounded-xl border bg-card">
              <img
                src="/home.jpg"
                alt="Prometheus alarm clock"
                className="w-full"
              />
              <div className="p-6 text-sm text-muted-foreground">
                A Raspberry Pi–powered alarm clock with a Nixie-tube display,
                vibration motor, and over-the-network configuration.
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Toaster richColors position="bottom-center" />
    </div>
  );
}
